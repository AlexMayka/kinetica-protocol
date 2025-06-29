package tcp

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

const (
	DefaultTransportCRC   = message.TransportNone
	DefaultMaxMessageSize = 64 * 1024
)

type ConnectionTCP struct {
	conn         net.Conn
	ctx          context.Context
	reader       *bufio.Reader
	writeTimeout time.Duration
	readTimeout  time.Duration
	packetID     atomic.Uint32
}

func NewConnectionTCP(conn net.Conn, ctx context.Context, writeTimeout, readTimeout time.Duration) transport.Connection {
	return &ConnectionTCP{
		conn:         conn,
		ctx:          ctx,
		reader:       bufio.NewReader(conn),
		writeTimeout: writeTimeout,
		readTimeout:  readTimeout,
		packetID:     atomic.Uint32{},
	}
}

func (c *ConnectionTCP) getNextPacketID() uint8 {
	id := c.packetID.Add(1)
	return uint8(id % 256)
}

func (c *ConnectionTCP) Send(msg message.Message, msgType message.MsgType) (transport.SendStatus, error) {
	if msg == nil {
		return transport.SendFailed, fmt.Errorf("%w: message is nil", transport.ErrInvalidMessageSize)
	}

	binaryMsg, err := codec.Marshal(msg, c.getNextPacketID(), msgType, DefaultTransportCRC)
	if err != nil {
		return transport.SendFailed, fmt.Errorf("%w: failed to marshal message: %w", transport.ErrSendFailed, err)
	}

	if len(binaryMsg) > DefaultMaxMessageSize && msgType != message.MsgTypeFragment {
		return transport.SendFailed, fmt.Errorf("%w: message size %d exceeds maximum %d bytes", transport.ErrMsgLarge, len(binaryMsg), DefaultMaxMessageSize)
	}

	select {
	case <-c.ctx.Done():
		return transport.SendFailed, fmt.Errorf("%w: %w", transport.ErrContextCanceled, c.ctx.Err())
	default:
	}

	if c.writeTimeout > 0 {
		if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
			return transport.SendFailed, fmt.Errorf("%w: failed to set write deadline: %w", transport.ErrSendFailed, err)
		}
	}

	n, err := c.conn.Write(binaryMsg)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return transport.SendFailed, fmt.Errorf("%w: %w", transport.ErrWriteTimeout, err)
		}
		if errors.Is(err, net.ErrClosed) {
			return transport.SendFailed, fmt.Errorf("%w: %w", transport.ErrConnectionClosed, err)
		}
		return transport.SendFailed, fmt.Errorf("%w: failed to write %d bytes: %w", transport.ErrSendFailed, len(binaryMsg), err)
	}

	if n != len(binaryMsg) {
		return transport.SendFailed, fmt.Errorf("%w: partial write: wrote %d of %d bytes", transport.ErrSendFailed, n, len(binaryMsg))
	}

	return transport.SendSuccess, nil
}

func (c *ConnectionTCP) Receive() (message.Message, error) {
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
	if int(payloadLength) > DefaultMaxMessageSize {
		return nil, fmt.Errorf("%w: payload length %d exceeds maximum %d", transport.ErrMsgLarge, payloadLength, DefaultMaxMessageSize)
	}

	footerSize := message.GetFooterSize(DefaultTransportCRC)
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
	msg, err := codec.Unmarshal(fullMessage, DefaultTransportCRC)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to unmarshal message: %w", transport.ErrReceiveFailed, err)
	}

	return msg, nil
}

func (c *ConnectionTCP) State() transport.ConnectionState {
	select {
	case <-c.ctx.Done():
		return transport.StateDisconnected
	default:
	}

	_, err := c.reader.Peek(1)
	if err == nil {
		return transport.StateConnect
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return transport.StateConnect
	}

	return transport.StateDisconnected
}

func (c *ConnectionTCP) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *ConnectionTCP) Close() error {
	return c.conn.Close()
}
