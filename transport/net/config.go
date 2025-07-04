// Package net provides TCP and UDP transport implementations for the Kinetica protocol.
// It supports both client and server modes with configurable timeouts and message size limits.
// TCP uses no CRC validation (relying on TCP's built-in reliability), while UDP uses 8-bit CRC.
package net

import "time"

// Config defines network transport configuration parameters for TCP and UDP connections.
type Config struct {
	Address      string        // Network address to bind/connect (e.g., ":8080", "192.168.1.100:8080")
	WriteTimeout time.Duration // Timeout for write operations (0 = no timeout)
	ReadTimeout  time.Duration // Timeout for read operations (0 = no timeout)
}
