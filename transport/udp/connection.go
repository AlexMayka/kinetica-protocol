package udp

import (
	"bufio"
	"context"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"net"
	"sync/atomic"
	"time"
)

const (
	DefaultTransportCRC   = message.TransportCRC16
	DefaultMaxMessageSize = 64 * 1024
)

type ConnectionUDP struct {
	conn         net.Conn
	ctx          context.Context
	reader       *bufio.Reader
	writeTimeout time.Duration
	readTimeout  time.Duration
	packetID     atomic.Uint32
}

func NewConnectionUDP(conn net.Conn, ctx context.Context, writeTimeout, readTimeout time.Duration) transport.Connection {
	return &ConnectionUDP{
		conn:         conn,
		ctx:          ctx,
		reader:       bufio.NewReader(conn),
		writeTimeout: writeTimeout,
		readTimeout:  readTimeout,
		packetID:     atomic.Uint32{},
	}
}

func (c *ConnectionUDP) Send(msg message.Message, msgType message.MsgType) (transport.SendStatus, error) {
	return transport.SendSuccess, nil
}

func (c *ConnectionUDP) Receive() (message.Message, error) {
	return nil, nil
}

func (c *ConnectionUDP) State() transport.ConnectionState {
	return transport.StateConnect
}

func (c *ConnectionUDP) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *ConnectionUDP) Close() error {
	return c.conn.Close()
}
