package net

import (
	"context"
	"errors"
	"fmt"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"net"
)

// TCP transport constants for protocol configuration.
const (
	TCPTransportCRC   = message.TransportNone // No CRC validation (TCP provides reliability)
	TCPMaxMessageSize = 64 * 1024             // Maximum message size (64KB) for TCP
)

// TCPTransport implements TCP transport layer for reliable streaming communication.
// It supports both client connections and server listening for multiple clients.
type TCPTransport struct {
	config   Config           // TCP connection configuration
	listener *net.TCPListener // TCP listener for server mode
	ctx      context.Context  // Context for lifecycle management
	cancel   context.CancelFunc // Cancel function for cleanup
}

// NewTCP creates a new TCP transport instance with the specified configuration.
func NewTCP(config Config) *TCPTransport {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPTransport{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connection establishes a TCP client connection to the configured address.
func (t *TCPTransport) Connection() (transport.Connection, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", t.config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address '%s': %w", t.config.Address, err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("connection timeout to %s: %w", tcpAddr.String(), err)
		}
		return nil, fmt.Errorf("failed to connect to %s: %w", tcpAddr.String(), err)
	}

	return NewConnection(conn, t.ctx, t.config.WriteTimeout, t.config.ReadTimeout, TCPTransportCRC, TCPMaxMessageSize), nil
}

// Listen starts a TCP server and returns a channel of incoming connections.
// Each accepted connection is wrapped in a protocol Connection interface.
func (t *TCPTransport) Listen() (<-chan transport.Connection, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", t.config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address '%s': %w", t.config.Address, err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", tcpAddr.String(), err)
	}

	t.listener = listener
	ch := make(chan transport.Connection)

	go func() {
		defer close(ch)
		for {
			conn, err := listener.Accept()
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if err != nil {
				continue
			}
			ch <- NewConnection(conn, t.ctx, t.config.WriteTimeout, t.config.ReadTimeout, TCPTransportCRC, TCPMaxMessageSize)
		}
	}()

	return ch, nil
}

// Close shuts down the TCP transport and stops accepting new connections.
func (t *TCPTransport) Close() error {
	t.cancel()
	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}
