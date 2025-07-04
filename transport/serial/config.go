// Package serial provides RS232/UART serial transport implementation for the Kinetica protocol.
// It supports communication with embedded devices like ESP32, Arduino, and other microcontrollers
// over serial ports with configurable baud rates, parity, and flow control settings.
package serial

import (
	s "go.bug.st/serial"
	"time"
)

// Config defines serial port configuration parameters for RS232/UART communication.
// It includes all standard serial communication settings plus protocol-specific timeouts.
type Config struct {
	Port string // Serial port device path (e.g., "/dev/ttyUSB0", "COM3")

	// Serial port communication parameters
	BaudRate             int       // Communication speed (e.g., 9600, 115200)
	DataBits             int       // Number of data bits (typically 8)
	Parity               s.Parity  // Parity checking (None, Even, Odd)
	StopBits             s.StopBits // Number of stop bits (1 or 2)
	InitialStatusBitsRTS bool      // Request To Send signal state
	InitialStatusBitsDTR bool      // Data Terminal Ready signal state

	// Protocol timeouts
	ReadTimeout time.Duration // Timeout for read operations
}
