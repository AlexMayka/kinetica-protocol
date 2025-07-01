package message

type MsgType int8
type Version uint8

var MagicBytes = [2]byte{'K', 'N'}

const (
	V1 Version = 1
	V2 Version = 2
)

const (
	MsgTypeCommand         MsgType = 0x01
	MsgTypeConfig          MsgType = 0x02
	MsgTypeHeartbeat       MsgType = 0x03
	MsgTypeSensorData      MsgType = 0x04
	MsgTypeCustom          MsgType = 0x05
	MsgTypeTimeSync        MsgType = 0x06
	MsgTypeAck             MsgType = 0x07
	MsgTypeRegister        MsgType = 0x08
	MsgTypeFragment        MsgType = 0x09
	MsgTypeRelayed         MsgType = 0x0A
	MsgTypeSensorDataMulti MsgType = 0x0B
)

const HeaderSize = 6

type Header struct {
	Magic    [2]byte
	PacketID uint8
	Version  Version
	Type     MsgType
	Length   uint8
}

func NewHeader(packetID uint8, msgType MsgType, payloadLength uint8) Header {
	return Header{
		Magic:    MagicBytes,
		PacketID: packetID,
		Version:  V1,
		Type:     msgType,
		Length:   payloadLength,
	}
}
