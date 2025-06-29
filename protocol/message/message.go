package message

type Message interface {
	MessageType() MsgType
}
