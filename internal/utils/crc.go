// Package utils provides internal utility functions for the Kinetica protocol implementation.
// This package contains CRC calculation functions used for data integrity verification
// across different transport layers.
package utils

// CRC32 calculates a 32-bit CRC checksum for the given data using the IEEE 802.3 polynomial.
// It implements the standard CRC-32 algorithm with polynomial 0x04C11DB7.
// Used for transport layers that require high data integrity verification.
func CRC32(data []byte) uint32 {
	var crc uint32 = 0xFFFFFFFF
	const poly uint32 = 0x04C11DB7

	for _, b := range data {
		crc ^= uint32(b) << 24
		for i := 0; i < 8; i++ {
			if crc&0x80000000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}

	return ^crc
}

// CRC16 calculates a 16-bit CRC checksum for the given data using the CCITT polynomial.
// It implements CRC-16-CCITT algorithm with polynomial 0x1021.
// Used for medium-reliability transport layers like serial communication.
func CRC16(data []byte) uint16 {
	var crc16 uint16 = 0xffff

	for _, v := range data {
		crc16 ^= uint16(v) << 8
		for i := 0; i < 8; i++ {
			if crc16&0x8000 != 0 {
				crc16 = (crc16 << 1) ^ 0x1021
			} else {
				crc16 <<= 1
			}
		}
	}

	return crc16
}

// CRC8 calculates an 8-bit CRC checksum for the given data using polynomial 0x07.
// It implements a simple CRC-8 algorithm suitable for short messages.
// Used for low-overhead transport layers like BLE where bandwidth is limited.
func CRC8(data []byte) uint8 {
	var crc uint8 = 0x00
	poly := uint8(0x07)

	for _, b := range data {
		crc ^= b
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
