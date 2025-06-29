package message

import (
	"kinetica-protocol/internal/utils"
)

type TransportCRC uint8

const (
	TransportCRC8   TransportCRC = 0x01
	TransportCRC16  TransportCRC = 0x02
	TransportCRC32  TransportCRC = 0x03
	TransportLength TransportCRC = 0x04
	TransportNone   TransportCRC = 0x05
)

type Footer struct {
	Bytes []byte
}

func NewFooter(transportType TransportCRC, data []byte) *Footer {
	var value []byte

	switch transportType {
	case TransportCRC8:
		value = CalculateChecksum8(data)
	case TransportCRC16:
		value = CalculateChecksum16(data)
	case TransportCRC32:
		value = CalculateChecksum32(data)
	case TransportLength:
		value = CalculateLength(data)
	case TransportNone:
		value = nil
	}

	return &Footer{Bytes: value}
}

func CalculateChecksum8(data []byte) []byte {
	return []byte{utils.CRC8(data)}
}

func CalculateChecksum16(data []byte) []byte {
	crc := utils.CRC16(data)
	return []byte{byte(crc), byte(crc >> 8)}
}

func CalculateChecksum32(data []byte) []byte {
	crc := utils.CRC32(data)
	return []byte{byte(crc), byte(crc >> 8), byte(crc >> 16), byte(crc >> 24)}
}

func CalculateLength(data []byte) []byte {
	return []byte{byte(len(data))}
}

func GetFooterSize(transport TransportCRC) int {
	switch transport {
	case TransportCRC8:
		return 1
	case TransportCRC16:
		return 2
	case TransportCRC32:
		return 4
	case TransportLength:
		return 1
	default:
		return 0
	}
}
