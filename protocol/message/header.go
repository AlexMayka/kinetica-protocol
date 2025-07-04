package message

// MsgType represents the type identifier for protocol messages.
type MsgType int8

// Version represents the protocol version number.
type Version uint8

// MagicBytes are the protocol magic bytes used to identify valid packets.
var MagicBytes = [2]byte{'K', 'N'}

// Protocol version constants.
const (
	V1 Version = 1 // Version 1 of the protocol
	V2 Version = 2 // Version 2 of the protocol (future use)
)

// Message type constants defining all supported message types in the protocol.
const (
	MsgTypeCommand         MsgType = 0x01 // Sensor command message
	MsgTypeConfig          MsgType = 0x02 // Configuration message
	MsgTypeHeartbeat       MsgType = 0x03 // Heartbeat/status message
	MsgTypeSensorData      MsgType = 0x04 // Single sensor data message
	MsgTypeCustom          MsgType = 0x05 // Custom data message
	MsgTypeTimeSync        MsgType = 0x06 // Time synchronization message
	MsgTypeAck             MsgType = 0x07 // Acknowledgment message
	MsgTypeRegister        MsgType = 0x08 // Sensor registration message
	MsgTypeFragment        MsgType = 0x09 // Fragmented message part
	MsgTypeRelayed         MsgType = 0x0A // Relayed message through hub
	MsgTypeSensorDataMulti MsgType = 0x0B // Multiple sensor data in one message
)

// HeaderSize is the fixed size of the protocol header in bytes.
const HeaderSize = 6

// Header represents the protocol message header containing packet metadata.
type Header struct {
	Magic    [2]byte // Protocol magic bytes ('K', 'N')
	PacketID uint8   // Unique packet identifier (0-255, wraps around)
	Version  Version // Protocol version
	Type     MsgType // Message type identifier
	Length   uint8   // Payload length in bytes
}

// NewHeader creates a new protocol header with the specified parameters.
// It automatically sets the magic bytes and protocol version.
func NewHeader(packetID uint8, msgType MsgType, payloadLength uint8) Header {
	return Header{
		Magic:    MagicBytes,
		PacketID: packetID,
		Version:  V1,
		Type:     msgType,
		Length:   payloadLength,
	}
}
