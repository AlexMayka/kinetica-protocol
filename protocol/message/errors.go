package message

import "errors"

// Protocol error definitions for message validation and processing.
var (
	ErrInvalidMagic       = errors.New("invalid magic bytes")          // Magic bytes don't match expected 'KN'
	ErrInvalidVersion     = errors.New("invalid protocol version")     // Unsupported protocol version
	ErrMessageTooLarge    = errors.New("message exceeds maximum size") // Message size exceeds transport limits
	ErrInvalidChecksum    = errors.New("invalid message checksum")     // CRC validation failed
	ErrUnknownMessageType = errors.New("unknown message type")         // Unrecognized message type identifier
)
