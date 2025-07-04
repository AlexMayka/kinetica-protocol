package serial

import (
	"context"
	"fmt"
	"go.bug.st/serial"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
)

// Serial transport constants for protocol configuration.
const (
	TransportCRC = message.TransportCRC8 // 8-bit CRC for serial communication reliability
	MaxMsgSize   = 4 * 1024              // Maximum message size (4KB) for serial buffers
)

// Transport implements serial/UART transport layer for embedded device communication.
// It manages serial port configuration and provides both client and server modes.
type Transport struct {
	config   Config             // Serial port configuration
	listener serial.Port        // Serial port for server mode
	mode     serial.Mode        // Serial communication mode settings
	ctx      context.Context    // Context for lifecycle management
	cancel   context.CancelFunc // Cancel function for cleanup
}

// NewSerial creates a new serial transport instance with the specified configuration.
// It configures serial port parameters including baud rate, parity, and control signals.
func NewSerial(c Config) transport.Transport {
	ctx, cancel := context.WithCancel(context.Background())
	return &Transport{
		config: c,
		mode: serial.Mode{
			BaudRate: c.BaudRate,
			DataBits: c.DataBits,
			Parity:   c.Parity,
			StopBits: c.StopBits,
			InitialStatusBits: &serial.ModemOutputBits{
				RTS: true,
				DTR: true,
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connection opens a serial port connection for client communication.
func (t *Transport) Connection() (transport.Connection, error) {
	port, err := serial.Open(t.config.Port, &t.mode)
	if err != nil {
		return nil, fmt.Errorf("%w: can't open Port: %v: %w", transport.ErrConn, t.config.Port, err)
	}
	return NewConnection(port, t.ctx, t.config.ReadTimeout, TransportCRC, MaxMsgSize), nil
}

// Listen opens a serial port for server mode and returns a single connection.
// Serial communication is point-to-point, so only one connection is provided.
func (t *Transport) Listen() (<-chan transport.Connection, error) {
	port, err := serial.Open(t.config.Port, &t.mode)
	if err != nil {
		return nil, fmt.Errorf("%w: can't open Port: %v: %w", transport.ErrConn, t.config.Port, err)
	}

	ch := make(chan transport.Connection)
	ch <- NewConnection(port, t.ctx, t.config.ReadTimeout, TransportCRC, MaxMsgSize)
	close(ch)

	return ch, nil
}

// Close shuts down the serial transport and releases the port.
func (t *Transport) Close() error {
	t.cancel()
	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}
