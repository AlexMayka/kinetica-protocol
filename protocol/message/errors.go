package message

import "errors"

var (
	ErrInvalidMagic       = errors.New("invalid magic bytes")
	ErrInvalidVersion     = errors.New("invalid protocol version")
	ErrMessageTooLarge    = errors.New("message exceeds maximum size")
	ErrInvalidChecksum    = errors.New("invalid message checksum")
	ErrUnknownMessageType = errors.New("unknown message type")
)
