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
	StateConnecting   ConnectionState = 0x02
	StateDisconnected ConnectionState = 0x03
	StateError        ConnectionState = 0x04
)

type Connection interface {
	Send(msg message.Message, crcType message.TransportCRC) (SendStatus, error)
	Receive() (message.Message, error)
	State() SendStatus
	RemoteAddr() net.Addr
}

type Trasport interface {
	Type() TypeTransport
	Connection(address net.Addr) (Connection, error)
	Listen(address string) (<-chan Connection, error)
	Close() error
}

type Config struct {
	Type           TypeTransport        `json:"type"`
	Address        string               `json:"address"`
	Timeout        time.Duration        `json:"timeout"`
	DefaultCRCType message.TransportCRC `json:"default_crc_type"`

	BaudRate int `json:"baud_rate,omitempty"`

	ClientID string `json:"client_id,omitempty"`
	Topic    string `json:"topic,omitempty"`

	RetryAttempts int           `json:"retry_attempts"`
	RetryDelay    time.Duration `json:"retry_delay"`
}
