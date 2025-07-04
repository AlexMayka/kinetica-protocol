package transport

import "kinetica-protocol/protocol/message"

// ConnectionState represents the current state of a transport connection.
type ConnectionState uint8

// Connection state constants.
const (
	StateConnected    ConnectionState = 0x01 // Connection is active and ready for communication
	StateDisconnected ConnectionState = 0x02 // Connection is closed or failed
)

// Connection represents an active communication channel between two endpoints.
// It provides methods for sending and receiving protocol messages with automatic
// encoding/decoding and appropriate CRC validation for the transport type.
type Connection interface {
	// Send encodes and transmits a protocol message over the connection.
	// The message is automatically encoded with appropriate headers and CRC
	// based on the transport type.
	Send(msg message.Message, msgType message.MsgType) error
	
	// Receive waits for and decodes an incoming protocol message.
	// Returns the decoded message or an error if reception/validation fails.
	// This method may block until a message arrives or timeout occurs.
	Receive() (message.Message, error)
	
	// State returns the current connection state.
	State() ConnectionState
	
	// Close gracefully terminates the connection and releases resources.
	// After calling Close, the connection should not be used for further communication.
	Close() error
}