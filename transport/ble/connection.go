package ble

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"kinetica-protocol/protocol/codec"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"sync/atomic"
	"time"
	"tinygo.org/x/bluetooth"
)

// bleReader adapts BLE notification channel to io.Reader interface for bufio compatibility.
// It manages fragmented BLE packets and provides timeout-based reading.
type bleReader struct {
	rxBuffer    chan []byte   // Channel receiving BLE notification data
	current     []byte        // Current data buffer being read
	pos         int           // Current position in the data buffer
	readTimeout time.Duration // Timeout for waiting for new data
}

// Read implements io.Reader interface for BLE notification data.
// It handles packet fragmentation and applies read timeouts.
func (r *bleReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.current) {
		select {
		case data, ok := <-r.rxBuffer:
			if !ok {
				return 0, io.EOF
			}
			r.current = data
			r.pos = 0
		case <-time.After(r.readTimeout):
			return 0, errors.New("read timeout")
		}
	}

	n = copy(p, r.current[r.pos:])
	r.pos += n
	return n, nil
}

// Connection represents an active BLE connection with GATT characteristics.
// It manages protocol message transmission over BLE write/notify characteristics.
type Connection struct {
	ctx         context.Context                    // Context for connection lifecycle
	device      *bluetooth.Device                  // Connected BLE device
	writeChar   bluetooth.DeviceCharacteristic     // Characteristic for sending data
	notifyChar  bluetooth.DeviceCharacteristic     // Characteristic for receiving notifications
	reader      *bufio.Reader                      // Buffered reader for protocol messages
	readTimeout time.Duration                      // Timeout for read operations
	packetID    atomic.Uint32                      // Atomic counter for unique packet IDs
	rxBuffer    chan []byte                        // Buffer for incoming notification data
}

// NewConnection creates a new BLE connection with the specified device and configuration.
// It sets up GATT characteristics and notification handlers for protocol communication.
func NewConnection(ctx context.Context, device *bluetooth.Device, config Config) (transport.Connection, error) {

	rxBuffer := make(chan []byte, 10)
	bleReader := &bleReader{
		rxBuffer:    rxBuffer,
		readTimeout: config.ReadTimeout,
	}

	conn := &Connection{
		ctx:         ctx,
		device:      device,
		reader:      bufio.NewReader(bleReader),
		readTimeout: config.ReadTimeout,
		packetID:    atomic.Uint32{},
		rxBuffer:    rxBuffer,
	}

	if err := conn.setupCharacteristics(config, rxBuffer); err != nil {
		return nil, fmt.Errorf("failed to setup characteristics: %v", err)
	}

	return conn, nil
}

// setupCharacteristics discovers and configures GATT service and characteristics.
// It enables notifications on the notify characteristic for receiving data.
func (c *Connection) setupCharacteristics(config Config, rxBuffer chan []byte) error {
	services, err := c.device.DiscoverServices([]bluetooth.UUID{config.ServiceUUID})
	if err != nil {
		return fmt.Errorf("failed to discover services: %v", err)
	}

	if len(services) == 0 {
		return fmt.Errorf("service not found")
	}

	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{
		config.WriteCharUUID,
		config.NotifyCharUUID,
	})
	if err != nil {
		return fmt.Errorf("failed to discover characteristics: %v", err)
	}

	for _, char := range chars {
		if char.UUID() == config.WriteCharUUID {
			c.writeChar = char
		} else if char.UUID() == config.NotifyCharUUID {
			c.notifyChar = char
			err := char.EnableNotifications(func(buf []byte) {
				select {
				case rxBuffer <- buf:
				case <-time.After(10 * time.Millisecond):
				}
			})
			if err != nil {
				return fmt.Errorf("failed to enable notifications: %v", err)
			}
		}
	}

	if c.writeChar.UUID() == (bluetooth.UUID{}) || c.notifyChar.UUID() == (bluetooth.UUID{}) {
		return fmt.Errorf("required characteristics not found")
	}

	return nil
}

// Send encodes and transmits a protocol message over the BLE write characteristic.
// It validates message size against BLE MTU constraints and handles fragmentation if needed.
func (c *Connection) Send(msg message.Message, msgType message.MsgType) error {
	if msg == nil {
		return fmt.Errorf("%w: message is nil", transport.ErrInvalidMessageSize)
	}

	binaryMsg, err := codec.Marshal(msg, c.getNextPacketID(), msgType, TransportCRC)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal message: %w", transport.ErrSendFailed, err)
	}

	if len(binaryMsg) > MaxMessageSize && msgType != message.MsgTypeFragment {
		return fmt.Errorf("%w: message size %d exceeds maximum %d bytes", transport.ErrMsgLarge, len(binaryMsg), MaxMessageSize)
	}

	select {
	case <-c.ctx.Done():
		return fmt.Errorf("%w: %w", transport.ErrContextCanceled, c.ctx.Err())
	default:
	}

	_, err = c.writeChar.WriteWithoutResponse(binaryMsg)
	if err != nil {
		return fmt.Errorf("%w: failed to write: %w", transport.ErrSendFailed, err)
	}

	return nil
}

// Receive reads and decodes a protocol message from BLE notifications.
// It handles packet fragmentation and validates the complete message.
func (c *Connection) Receive() (message.Message, error) {
	select {
	case <-c.ctx.Done():
		return nil, fmt.Errorf("%w: %w", transport.ErrContextCanceled, c.ctx.Err())
	default:
	}

	headerBuf := make([]byte, message.HeaderSize)
	n, err := io.ReadFull(c.reader, headerBuf)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("%w: connection closed by peer", transport.ErrConnectionClosed)
		}
		return nil, fmt.Errorf("%w: failed to read header (%d/%d bytes): %w", transport.ErrReceiveFailed, n, message.HeaderSize, err)
	}

	payloadLength := headerBuf[5]
	if int(payloadLength) > MaxMessageSize {
		return nil, fmt.Errorf("%w: payload length %d exceeds maximum %d", transport.ErrMsgLarge, payloadLength, MaxMessageSize)
	}

	footerSize := message.GetFooterSize(TransportCRC)
	remainingSize := int(payloadLength) + footerSize
	remainingBuf := make([]byte, remainingSize)

	n, err = io.ReadFull(c.reader, remainingBuf)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("%w: connection closed while reading payload", transport.ErrConnectionClosed)
		}
		return nil, fmt.Errorf("%w: failed to read payload (%d/%d bytes): %w", transport.ErrReceiveFailed, n, remainingSize, err)
	}

	fullMessage := append(headerBuf, remainingBuf...)
	msg, err := codec.Unmarshal(fullMessage, TransportCRC)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal message: %w", transport.ErrReceiveFailed, err)
	}

	return msg, nil
}

// State returns the current connection state based on context status.
func (c *Connection) State() transport.ConnectionState {
	select {
	case <-c.ctx.Done():
		return transport.StateDisconnected
	default:
		return transport.StateConnected
	}
}

// Close gracefully terminates the BLE connection and cleans up resources.
// It disables notifications, closes buffers, and disconnects from the device.
func (c *Connection) Close() error {
	if c.notifyChar.UUID() != (bluetooth.UUID{}) {
		_ = c.notifyChar.EnableNotifications(nil)
	}

	if c.rxBuffer != nil {
		close(c.rxBuffer)
	}

	return c.device.Disconnect()
}

// getNextPacketID generates a unique packet ID using atomic increment with wraparound.
func (c *Connection) getNextPacketID() uint8 {
	return uint8(c.packetID.Add(1) % 256)
}
