package codec

import (
	"errors"
)

var (
	ErrEncodingFailed     = errors.New("binary encoding failed")
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrBufferOverflow     = errors.New("buffer overflow")

	ErrDecodingFailed     = errors.New("binary decoding failed")
	ErrInsufficientData   = errors.New("insufficient data to decode")
	ErrInvalidMagicBytes  = errors.New("invalid magic bytes")
	ErrUnsupportedVersion = errors.New("unsupported protocol version")
	ErrInvalidFooter      = errors.New("footer validation failed")
	ErrInvalidCRC         = errors.New("CRC validation failed")
	ErrPayloadTooLarge    = errors.New("payload exceeds maximum size")
	ErrMessageTooShort    = errors.New("message too short")
	ErrInvalidHeader      = errors.New("invalid header")
	ErrUnknownMessageType = errors.New("unknown message type")
)
