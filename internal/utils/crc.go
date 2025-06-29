package utils

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
