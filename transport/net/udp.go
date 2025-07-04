package net

import (
	"context"
	"fmt"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"net"
)

// UDP transport constants for protocol configuration.
const (
	UDPTransportCRC   = message.TransportCRC8 // 8-bit CRC for UDP datagram validation
	UDPMaxMessageSize = 1472                  // Maximum UDP payload size (Ethernet MTU - headers)
)

// UDPTransport implements UDP transport layer for datagram-based communication.
// It provides connectionless communication with CRC validation for message integrity.
type UDPTransport struct {
	config Config           // UDP connection configuration
	conn   *net.UDPConn     // UDP connection for server mode
	ctx    context.Context  // Context for lifecycle management
	cancel context.CancelFunc // Cancel function for cleanup
}

// NewUDP creates a new UDP transport instance with the specified configuration.
func NewUDP(config Config) *UDPTransport {
	ctx, cancel := context.WithCancel(context.Background())
	return &UDPTransport{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connection establishes a UDP client connection to the configured address.
func (t *UDPTransport) Connection() (transport.Connection, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", t.config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address '%s': %w", t.config.Address, err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", udpAddr.String(), err)
	}

	return NewConnection(conn, t.ctx, t.config.WriteTimeout, t.config.ReadTimeout, UDPTransportCRC, UDPMaxMessageSize), nil
}

// Listen creates a UDP server socket and returns a single connection for all clients.
// Unlike TCP, UDP uses a single connection that handles all client communications.
func (t *UDPTransport) Listen() (<-chan transport.Connection, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", t.config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address '%s': %w", t.config.Address, err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", udpAddr.String(), err)
	}

	t.conn = conn
	ch := make(chan transport.Connection, 1)
	ch <- NewConnection(conn, t.ctx, t.config.WriteTimeout, t.config.ReadTimeout, UDPTransportCRC, UDPMaxMessageSize)
	close(ch)

	return ch, nil
}

// Close shuts down the UDP transport and releases the socket.
func (t *UDPTransport) Close() error {
	t.cancel()
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}