// Package message defines the core protocol message types, constants, and interfaces
// for the Kinetica protocol. It provides structures for all supported message types
// including sensor data, commands, heartbeats, and configuration messages.
package message

// Message represents any protocol message that can be encoded/decoded.
// All message types must implement this interface to specify their type identifier.
type Message interface {
	MessageType() MsgType
}
