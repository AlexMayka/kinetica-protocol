package net

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"kinetica-protocol/protocol/codec"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"net"
	"sync/atomic"
	"time"
)

// Connection represents a network connection (TCP or UDP) with protocol message handling.
// It manages timeouts, CRC validation, and message size constraints for the transport.
type Connection struct {
	conn           net.Conn                 // Underlying network connection
	ctx            context.Context          // Context for operation cancellation
	reader         *bufio.Reader            // Buffered reader for efficient message parsing
	writeTimeout   time.Duration            // Timeout for write operations
	readTimeout    time.Duration            // Timeout for read operations
	packetID       atomic.Uint32            // Atomic counter for unique packet IDs
	transportCRC   message.TransportCRC     // CRC type for this transport
	maxMessageSize int                      // Maximum message size for this transport
}

// NewConnection creates a new network connection wrapper with protocol support.
// It configures timeouts, CRC validation, and message size limits for the connection.
func NewConnection(conn net.Conn, ctx context.Context, writeTimeout, readTimeout time.Duration, transportCRC message.TransportCRC, maxMessageSize int) *Connection {
	return &Connection{
		conn:           conn,
		ctx:            ctx,
		reader:         bufio.NewReader(conn),
		writeTimeout:   writeTimeout,
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

// Send encodes and transmits a protocol message over the network connection.
// It handles timeouts, validates message size, and manages partial writes.
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

	if c.writeTimeout > 0 {
		if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
			return fmt.Errorf("%w: failed to set write deadline: %w", transport.ErrSendFailed, err)
		}
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

// Receive reads and decodes a protocol message from the network connection.
// It handles timeouts, validates message integrity, and manages partial reads.
func (c *Connection) Receive() (message.Message, error) {
	select {
	case <-c.ctx.Done():
		return nil, fmt.Errorf("%w: %w", transport.ErrContextCanceled, c.ctx.Err())
	default:
	}

	if c.readTimeout > 0 {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout)); err != nil {
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

// State returns the current connection state by checking context and connection health.
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

// RemoteAddr returns the remote network address of the connection.
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// LocalAddr returns the local network address of the connection.
func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// Stats returns connection statistics (currently not implemented).
func (c *Connection) Stats() (bytesSent, bytesReceived, messagesSent, messagesReceived uint64) {
	return 0, 0, 0, 0
}

// Close terminates the network connection and releases resources.
func (c *Connection) Close() error {
	return c.conn.Close()
}
