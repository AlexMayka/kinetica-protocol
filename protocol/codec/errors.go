package codec

import (
	"errors"
)

// Codec error definitions for encoding and decoding operations.
var (
	// Encoding errors
	ErrEncodingFailed     = errors.New("binary encoding failed")     // Binary serialization failed
	ErrInvalidMessageType = errors.New("invalid message type")       // Unknown message type for encoding
	ErrBufferOverflow     = errors.New("buffer overflow")            // Buffer capacity exceeded during encoding

	// Decoding errors
	ErrDecodingFailed     = errors.New("binary decoding failed")     // Binary deserialization failed
	ErrInsufficientData   = errors.New("insufficient data to decode") // Not enough bytes for complete decode
	ErrInvalidMagicBytes  = errors.New("invalid magic bytes")        // Magic bytes don't match protocol
	ErrUnsupportedVersion = errors.New("unsupported protocol version") // Unknown protocol version
	ErrInvalidFooter      = errors.New("footer validation failed")   // Footer CRC validation failed
	ErrInvalidCRC         = errors.New("CRC validation failed")      // CRC checksum mismatch
	ErrPayloadTooLarge    = errors.New("payload exceeds maximum size") // Payload size exceeds limits
	ErrMessageTooShort    = errors.New("message too short")          // Message shorter than minimum size
	ErrInvalidHeader      = errors.New("invalid header")             // Header structure is invalid
	ErrUnknownMessageType = errors.New("unknown message type")       // Unrecognized message type for decoding
)
