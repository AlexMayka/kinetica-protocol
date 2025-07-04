package message

import (
	"kinetica-protocol/internal/utils"
)

// TransportCRC defines the type of CRC validation used by different transport layers.
type TransportCRC uint8

// Transport CRC type constants for different levels of data integrity validation.
const (
	TransportCRC8   TransportCRC = 0x01 // 8-bit CRC for low-overhead transports (BLE)
	TransportCRC16  TransportCRC = 0x02 // 16-bit CRC for medium reliability (Serial)
	TransportCRC32  TransportCRC = 0x03 // 32-bit CRC for high reliability (TCP)
	TransportLength TransportCRC = 0x04 // Simple length validation (UDP)
	TransportNone   TransportCRC = 0x05 // No validation required
)

// Footer represents the protocol message footer containing validation data.
type Footer struct {
	Bytes []byte // CRC or validation bytes based on transport type
}

// NewFooter creates a new protocol footer with appropriate validation data
// based on the specified transport CRC type.
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

// CalculateChecksum8 computes an 8-bit CRC checksum for the given data.
// Used by BLE and other low-overhead transport layers.
func CalculateChecksum8(data []byte) []byte {
	return []byte{utils.CRC8(data)}
}

// CalculateChecksum16 computes a 16-bit CRC checksum for the given data.
// Used by serial and other medium-reliability transport layers.
func CalculateChecksum16(data []byte) []byte {
	crc := utils.CRC16(data)
	return []byte{byte(crc), byte(crc >> 8)}
}

// CalculateChecksum32 computes a 32-bit CRC checksum for the given data.
// Used by TCP and other high-reliability transport layers.
func CalculateChecksum32(data []byte) []byte {
	crc := utils.CRC32(data)
	return []byte{byte(crc), byte(crc >> 8), byte(crc >> 16), byte(crc >> 24)}
}

// CalculateLength creates a simple length validation footer.
// Used by UDP and other datagram-based transport layers.
func CalculateLength(data []byte) []byte {
	return []byte{byte(len(data))}
}

// GetFooterSize returns the footer size in bytes for the specified transport CRC type.
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
