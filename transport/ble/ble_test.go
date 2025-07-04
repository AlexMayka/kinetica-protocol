package ble

import (
	"testing"
	"time"
	"tinygo.org/x/bluetooth"
)

func TestTransport_NewBLE(t *testing.T) {
	config := Config{
		ServiceUUID:    bluetooth.New16BitUUID(0x180F),
		WriteCharUUID:  bluetooth.New16BitUUID(0x2A19),
		NotifyCharUUID: bluetooth.New16BitUUID(0x2A1A),
		ScanTimeout:    5 * time.Second,
		ReadTimeout:    1 * time.Second,
	}

	transport := NewBLE(config)
	if transport == nil {
		t.Fatal("Expected non-nil transport")
	}

	bleTransport, ok := transport.(*Transport)
	if !ok {
		t.Fatal("Expected *Transport type")
	}

	if bleTransport.adapter == nil {
		t.Error("Expected non-nil adapter")
	}

	if bleTransport.ctx == nil {
		t.Error("Expected non-nil context")
	}

	if bleTransport.cancel == nil {
		t.Error("Expected non-nil cancel function")
	}
}

func TestTransport_Listen(t *testing.T) {
	config := Config{
		ScanTimeout: 1 * time.Second,
	}

	transport := NewBLE(config)
	
	_, err := transport.Listen()
	if err == nil {
		t.Fatal("Expected error for Listen()")
	}

	if err.Error() != "unrealized method: BLE transport only supports client mode" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestTransport_Close(t *testing.T) {
	config := Config{
		ScanTimeout: 1 * time.Second,
	}

	transport := NewBLE(config)
	bleTransport := transport.(*Transport)

	err := transport.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	select {
	case <-bleTransport.ctx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Error("Context not cancelled after Close()")
	}
}

func TestTransport_ScanTimeout(t *testing.T) {
	mac := "AA:BB:CC:DD:EE:FF"
	config := Config{
		DeviceAddress: &mac,
		ScanTimeout:   100 * time.Millisecond,
		ReadTimeout:   1 * time.Second,
	}

	transport := NewBLE(config)
	
	start := time.Now()
	_, err := transport.Connection()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected error due to device not found")
	}

	if elapsed < 100*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Expected timeout around 100ms, got %v", elapsed)
	}
}

func TestTransport_ConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "with device name",
			config: Config{
				DeviceName:     stringPtr("MyDevice"),
				ServiceUUID:    bluetooth.New16BitUUID(0x180F),
				WriteCharUUID:  bluetooth.New16BitUUID(0x2A19),
				NotifyCharUUID: bluetooth.New16BitUUID(0x2A1A),
				ScanTimeout:    5 * time.Second,
			},
		},
		{
			name: "with device address",
			config: Config{
				DeviceAddress:  stringPtr("AA:BB:CC:DD:EE:FF"),
				ServiceUUID:    bluetooth.New16BitUUID(0x180F),
				WriteCharUUID:  bluetooth.New16BitUUID(0x2A19),
				NotifyCharUUID: bluetooth.New16BitUUID(0x2A1A),
				ScanTimeout:    5 * time.Second,
			},
		},
		{
			name: "with device filter",
			config: Config{
				DeviceFilter: func(name string) bool {
					return name == "ESP32"
				},
				ServiceUUID:    bluetooth.New16BitUUID(0x180F),
				WriteCharUUID:  bluetooth.New16BitUUID(0x2A19),
				NotifyCharUUID: bluetooth.New16BitUUID(0x2A1A),
				ScanTimeout:    5 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewBLE(tt.config)
			if transport == nil {
				t.Fatal("Expected non-nil transport")
			}
		})
	}
}

func TestTransport_ContextCancellation(t *testing.T) {
	config := Config{
		DeviceAddress: stringPtr("AA:BB:CC:DD:EE:FF"),
		ScanTimeout:   5 * time.Second,
	}

	transport := NewBLE(config)
	
	go func() {
		time.Sleep(50 * time.Millisecond)
		transport.Close()
	}()

	start := time.Now()
	_, err := transport.Connection()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected error due to context cancellation")
	}

	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected quick cancellation, got %v", elapsed)
	}
}

func stringPtr(s string) *string {
	return &s
}