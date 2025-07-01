package transport

import (
	"kinetica-protocol/protocol/message"
	"net"
	"time"
)

type TypeTransport uint8
type SendStatus uint8
type ConnectionState uint8

const (
	TCP    TypeTransport = 0x01
	UDP    TypeTransport = 0x02
	BLE    TypeTransport = 0x03
	SERIAL TypeTransport = 0x04
	MQTT   TypeTransport = 0x05
)

const (
	SendSuccess SendStatus = 0x01
	SendPending SendStatus = 0x02
	SendFailed  SendStatus = 0x03
	SendTimeout SendStatus = 0x04
)

const (
	StateConnect      ConnectionState = 0x01
	StateDisconnected ConnectionState = 0x03
)

type Config struct {
	Type           TypeTransport        `json:"type"`
	Address        string               `json:"address"`
	WriteTimeout   time.Duration        `json:"write_timeout"`
	ReadTimeout    time.Duration        `json:"read_timeout"`
	DefaultCRCType message.TransportCRC `json:"default_crc_type"`

	BaudRate int `json:"baud_rate,omitempty"`

	ClientID string `json:"client_id,omitempty"`
	Topic    string `json:"topic,omitempty"`

	RetryAttempts int           `json:"retry_attempts"`
	RetryDelay    time.Duration `json:"retry_delay"`
}

type Connection interface {
	Send(msg message.Message, msgType message.MsgType) (SendStatus, error)
	Receive() (message.Message, error)
	State() ConnectionState
	RemoteAddr() net.Addr
	Close() error
}

type Transport interface {
	Type() TypeTransport
	Connection() (Connection, error)
	Listen() (<-chan Connection, error)
	Close() error
}
