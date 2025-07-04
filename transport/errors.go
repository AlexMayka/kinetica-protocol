package transport

import "errors"

// Transport layer error definitions for connection and communication failures.
var (
	ErrInvalidMessageSize = errors.New("invalid message size")    // Message size validation failed
	ErrSendFailed         = errors.New("send failed")             // Message transmission failed
	ErrReceiveFailed      = errors.New("receive failed")          // Message reception failed
	ErrMsgLarge           = errors.New("message too large")       // Message exceeds transport limits
	ErrContextCanceled    = errors.New("context canceled")        // Operation canceled by context
	ErrWriteTimeout       = errors.New("write timeout")           // Send operation timed out
	ErrReadTimeout        = errors.New("read timeout")            // Receive operation timed out
	ErrConnectionClosed   = errors.New("connection closed")       // Connection was terminated
	ErrConn               = errors.New("can't connection")        // Connection establishment failed
	ErrUnrealizedMethod   = errors.New("unrealized method")       // Method not implemented by transport
)
