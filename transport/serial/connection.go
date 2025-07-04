package serial

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"go.bug.st/serial"
	"io"
	"kinetica-protocol/protocol/codec"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"net"
	"sync/atomic"
	"time"
)

// Connection represents an active serial port connection with protocol message handling.
// It manages serial communication with timeouts, CRC validation, and message size constraints.
type Connection struct {
	conn           serial.Port              // Serial port interface
	ctx            context.Context          // Context for operation cancellation
	reader         *bufio.Reader            // Buffered reader for efficient message parsing
	readTimeout    time.Duration            // Timeout for read operations
	packetID       atomic.Uint32            // Atomic counter for unique packet IDs
	transportCRC   message.TransportCRC     // CRC type for this transport
	maxMessageSize int                      // Maximum message size for this transport
}

// NewConnection creates a new serial connection wrapper with protocol support.
// It configures timeouts, CRC validation, and message size limits for the serial port.
func NewConnection(conn serial.Port, ctx context.Context, readTimeout time.Duration, transportCRC message.TransportCRC, maxMessageSize int) *Connection {
	return &Connection{
		conn:           conn,
		ctx:            ctx,
		reader:         bufio.NewReader(conn),
		readTimeout:    readTimeout,
		packetID:       atomic.Uint32{},
		transportCRC:   transportCRC,
		maxMessageSize: maxMessageSize,
	}
}

// getNextPacketID generates a unique packet ID using atomic increment with wraparound.
func (c *Connection) getNextPacketID() uint8 {
	id := c.packetID.Add(1)
	return uint8(id % 256)
}

// Send encodes and transmits a protocol message over the serial connection.
// It validates message size and handles partial writes for serial communication.
func (c *Connection) Send(msg message.Message, msgType message.MsgType) error {
	if msg == nil {
		return fmt.Errorf("%w: message is nil", transport.ErrInvalidMessageSize)
	}

	binaryMsg, err := codec.Marshal(msg, c.getNextPacketID(), msgType, c.transportCRC)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal message: %w", transport.ErrSendFailed, err)
	}

	if len(binaryMsg) > c.maxMessageSize && msgType != message.MsgTypeFragment {
		return fmt.Errorf("%w: message size %d exceeds maximum %d bytes", transport.ErrMsgLarge, len(binaryMsg), c.maxMessageSize)
	}

	select {
	case <-c.ctx.Done():
		return fmt.Errorf("%w: %w", transport.ErrContextCanceled, c.ctx.Err())
	default:
	}

	n, err := c.conn.Write(binaryMsg)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return fmt.Errorf("%w: %w", transport.ErrWriteTimeout, err)
		}
		if errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("%w: %w", transport.ErrConnectionClosed, err)
		}
		return fmt.Errorf("%w: failed to write %d bytes: %w", transport.ErrSendFailed, len(binaryMsg), err)
	}

	if n != len(binaryMsg) {
		return fmt.Errorf("%w: partial write: wrote %d of %d bytes", transport.ErrSendFailed, n, len(binaryMsg))
	}

	return nil
}

// Receive reads and decodes a protocol message from the serial connection.
// It handles timeouts and validates message integrity with CRC checking.
func (c *Connection) Receive() (message.Message, error) {
	select {
	case <-c.ctx.Done():
		return nil, fmt.Errorf("%w: %w", transport.ErrContextCanceled, c.ctx.Err())
	default:
	}

	if c.readTimeout > 0 {
		if err := c.conn.SetReadTimeout(c.readTimeout); err != nil {
			return nil, fmt.Errorf("%w: failed to set read deadline: %w", transport.ErrReceiveFailed, err)
		}
	}

	headerBuf := make([]byte, message.HeaderSize)
	n, err := io.ReadFull(c.reader, headerBuf)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("%w: connection closed by peer", transport.ErrConnectionClosed)
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("%w: %w", transport.ErrReadTimeout, err)
		}
		return nil, fmt.Errorf("%w: failed to read header (%d/%d bytes): %w", transport.ErrReceiveFailed, n, message.HeaderSize, err)
	}

	payloadLength := headerBuf[5]
	if int(payloadLength) > c.maxMessageSize {
		return nil, fmt.Errorf("%w: payload length %d exceeds maximum %d", transport.ErrMsgLarge, payloadLength, c.maxMessageSize)
	}

	footerSize := message.GetFooterSize(c.transportCRC)
	remainingSize := int(payloadLength) + footerSize
	remainingBuf := make([]byte, remainingSize)

	n, err = io.ReadFull(c.reader, remainingBuf)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("%w: connection closed while reading payload", transport.ErrConnectionClosed)
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("%w: %w", transport.ErrReadTimeout, err)
		}
		return nil, fmt.Errorf("%w: failed to read payload (%d/%d bytes): %w", transport.ErrReceiveFailed, n, remainingSize, err)
	}

	fullMessage := append(headerBuf, remainingBuf...)
	msg, err := codec.Unmarshal(fullMessage, c.transportCRC)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal message: %w", transport.ErrReceiveFailed, err)
	}

	return msg, nil
}

// State returns the current connection state by checking context and port status.
func (c *Connection) State() transport.ConnectionState {
	select {
	case <-c.ctx.Done():
		return transport.StateDisconnected
	default:
	}

	_, err := c.reader.Peek(1)
	if err == nil {
		return transport.StateConnected
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return transport.StateConnected
	}

	return transport.StateDisconnected
}

// Close terminates the serial connection and releases the port.
func (c *Connection) Close() error {
	return c.conn.Close()
}
