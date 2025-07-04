// Package transport provides abstraction layer for different communication protocols
// including TCP, UDP, Serial/UART, and BLE. It defines common interfaces for
// establishing connections and managing data transmission across various transport layers.
package transport

// Transport represents a communication transport layer that can establish connections
// or listen for incoming connections. Each transport implementation handles the
// specific protocol details while providing a unified interface.
type Transport interface {
	// Connection establishes a client connection to a remote endpoint.
	// Returns a Connection interface for sending/receiving messages.
	Connection() (Connection, error)
	
	// Listen starts listening for incoming connections on the transport.
	// Returns a channel that delivers new Connection instances as they arrive.
	// Used by server implementations to accept multiple clients.
	Listen() (<-chan Connection, error)
	
	// Close shuts down the transport and releases all resources.
	// Any active connections should be gracefully terminated.
	Close() error
}
