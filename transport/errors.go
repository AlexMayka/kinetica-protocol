package transport

import "errors"

var (
	ErrInvalidTypeTrans = errors.New("invalid type transport")
	ErrInvalidAddrTrans = errors.New("invalid address transport")
	ErrInvalidConfig    = errors.New("invalid transport configuration")

	ErrConnectionFailed  = errors.New("connection failed")
	ErrConnectionClosed  = errors.New("connection closed")
	ErrConnectionTimeout = errors.New("connection timeout")
	ErrListenAccept      = errors.New("listen accept failed")
	ErrClient            = errors.New("client error")

	ErrSendFailed         = errors.New("send failed")
	ErrReceiveFailed      = errors.New("receive failed")
	ErrMsgLarge           = errors.New("message too large")
	ErrMsgTooShort        = errors.New("message too short")
	ErrInvalidMessageSize = errors.New("invalid message size")

	ErrProtocolViolation  = errors.New("protocol violation")
	ErrInvalidPacketID    = errors.New("invalid packet ID")
	ErrFragmentationError = errors.New("fragmentation error")

	ErrWriteTimeout = errors.New("write timeout")
	ErrReadTimeout  = errors.New("read timeout")

	ErrContextCanceled  = errors.New("context canceled")
	ErrOperationAborted = errors.New("operation aborted")

	ErrSendError = ErrSendFailed
)
