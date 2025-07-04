package ble

import (
	"context"
	"fmt"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"tinygo.org/x/bluetooth"
)

// BLE transport constants for protocol configuration.
const (
	TransportCRC   = message.TransportCRC8 // 8-bit CRC for low-overhead BLE communication
	MaxMessageSize = 255                   // Maximum message size in bytes for BLE MTU constraints
)

// Transport implements the BLE transport layer for client connections.
// It manages BLE adapter, device discovery, and connection establishment.
type Transport struct {
	config  Config             // BLE connection configuration
	adapter *bluetooth.Adapter // Default BLE adapter for communication
	ctx     context.Context    // Context for cancellation and timeout control
	cancel  context.CancelFunc // Cancel function for context cleanup
}

// NewBLE creates a new BLE transport instance with the specified configuration.
// It initializes the default BLE adapter and context for operation management.
func NewBLE(c Config) transport.Transport {
	ctx, cancel := context.WithCancel(context.Background())

	return &Transport{
		config:  c,
		adapter: bluetooth.DefaultAdapter,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Listen is not supported by BLE transport as it operates in client-only mode.
// BLE devices typically act as peripherals that clients connect to.
func (t *Transport) Listen() (<-chan transport.Connection, error) {
	return nil, fmt.Errorf("%w: BLE transport only supports client mode", transport.ErrUnrealizedMethod)
}

// Connection establishes a BLE connection to the configured target device.
// It enables the BLE adapter, scans for the target device, and creates a connection.
func (t *Transport) Connection() (transport.Connection, error) {
	if err := t.adapter.Enable(); err != nil {
		return nil, fmt.Errorf("%w: failed to enable BLE adapter: %v", transport.ErrConn, err)
	}

	device, err := t.scanForDevice()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to find device: %v", transport.ErrConn, err)
	}

	return NewConnection(t.ctx, device, t.config)
}

// scanForDevice performs BLE device discovery based on configuration criteria.
// It scans for devices matching name, address, or custom filter with timeout.
func (t *Transport) scanForDevice() (*bluetooth.Device, error) {
	var foundDevice bluetooth.Device

	ctx, cancel := context.WithTimeout(context.Background(), t.config.ScanTimeout)
	defer cancel()

	ch := make(chan error)
	go func() {
		var connectErr error

		connect := func(adapter *bluetooth.Adapter, addr bluetooth.Address, conParam bluetooth.ConnectionParams) {
			device, err := adapter.Connect(addr, conParam)
			if err != nil {
				connectErr = fmt.Errorf("%w: failed to connect to device: %v", transport.ErrConn, err)
				_ = adapter.StopScan()
				return
			}

			foundDevice = device
			_ = adapter.StopScan()
		}

		err := t.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			if t.config.DeviceName != nil && (*t.config.DeviceName == result.LocalName()) {
				connect(adapter, result.Address, bluetooth.ConnectionParams{})
				return
			}

			if t.config.DeviceAddress != nil && (*t.config.DeviceAddress == result.Address.String()) {
				connect(adapter, result.Address, bluetooth.ConnectionParams{})
				return
			}

			if t.config.DeviceFilter != nil && t.config.DeviceFilter(result.LocalName()) {
				connect(adapter, result.Address, bluetooth.ConnectionParams{})
				return
			}
		})

		if connectErr != nil {
			ch <- connectErr
			return
		}

		ch <- err
	}()

	select {
	case err := <-ch:
		return &foundDevice, err
	case <-ctx.Done():
		_ = t.adapter.StopScan()
		return &foundDevice, fmt.Errorf("%w: scan timeout", transport.ErrConn)
	}
}

// Close shuts down the BLE transport and cancels any ongoing operations.
func (t *Transport) Close() error {
	t.cancel()
	return nil
}
