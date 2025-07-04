// Package ble provides Bluetooth Low Energy (BLE) transport implementation for the Kinetica protocol.
// It supports client-only connections to ESP32 and other BLE GATT devices using tinygo bluetooth library.
// This transport is optimized for low-power, low-bandwidth communication with 8-bit CRC validation.
package ble

import (
	"time"
	"tinygo.org/x/bluetooth"
)

// Config defines the configuration parameters for BLE transport connections.
// It specifies device discovery criteria, GATT service/characteristic UUIDs, and timing parameters.
type Config struct {
	// Device discovery criteria (exactly one should be specified)
	DeviceName    *string              // Target device name for connection
	DeviceAddress *string              // Target device MAC address for connection  
	DeviceFilter  func(string) bool    // Custom filter function for device selection

	// GATT service and characteristic UUIDs for protocol communication
	ServiceUUID    bluetooth.UUID // BLE service UUID hosting the protocol characteristics
	WriteCharUUID  bluetooth.UUID // Characteristic UUID for sending data to device
	NotifyCharUUID bluetooth.UUID // Characteristic UUID for receiving notifications from device

	// Timing configuration
	ScanTimeout time.Duration // Maximum time to scan for target device
	ReadTimeout time.Duration // Timeout for receiving data from device
}
