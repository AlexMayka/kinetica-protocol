package serial

import (
	"go.bug.st/serial"
	"testing"
	"time"
)

func TestSerialTransport_NewSerial(t *testing.T) {
	config := Config{
		Port:        "/dev/ttyUSB0",
		BaudRate:    115200,
		DataBits:    8,
		Parity:      serial.NoParity,
		StopBits:    serial.OneStopBit,
		ReadTimeout: 5 * time.Second,
	}

	transport := NewSerial(config)

	if transport == nil {
		t.Fatal("Expected non-nil transport")
	}

	st, ok := transport.(*Transport)
	if !ok {
		t.Fatal("Expected *Transport type")
	}

	if st.config.Port != "/dev/ttyUSB0" {
		t.Errorf("Expected Port /dev/ttyUSB0, got %s", st.config.Port)
	}

	if st.mode.BaudRate != 115200 {
		t.Errorf("Expected baud rate 115200, got %d", st.mode.BaudRate)
	}
}

func TestSerialTransport_ConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		port    string
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				BaudRate: 9600,
				DataBits: 8,
				Parity:   serial.NoParity,
				StopBits: serial.OneStopBit,
			},
			port:    "/dev/ttyS0",
			wantErr: false,
		},
		{
			name: "different baud rate",
			config: Config{
				Port:     "/dev/ttyUSB0",
				BaudRate: 115200,
				DataBits: 8,
				Parity:   serial.NoParity,
				StopBits: serial.OneStopBit,
			},
			wantErr: false,
		},
		{
			name: "with parity",
			config: Config{
				Port:     "/dev/ttyACM0",
				BaudRate: 19200,
				DataBits: 8,
				Parity:   serial.EvenParity,
				StopBits: serial.OneStopBit,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewSerial(tt.config)
			if transport == nil {
				t.Fatal("Expected non-nil transport")
			}
		})
	}
}

func TestSerialTransport_Close(t *testing.T) {
	config := Config{
		Port:     "/dev/ttyUSB0",
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	transport := NewSerial(config)
	st := transport.(*Transport)

	err := st.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	select {
	case <-st.ctx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Error("Context not cancelled after Close()")
	}
}

func TestSerialTransport_InitialStatusBits(t *testing.T) {
	config := Config{
		Port:                 "/dev/ttyUSB0",
		BaudRate:             115200,
		DataBits:             8,
		Parity:               serial.NoParity,
		StopBits:             serial.OneStopBit,
		InitialStatusBitsRTS: true,
		InitialStatusBitsDTR: true,
	}

	transport := NewSerial(config)
	st := transport.(*Transport)

	if st.mode.InitialStatusBits == nil {
		t.Fatal("InitialStatusBits should not be nil")
	}

	if !st.mode.InitialStatusBits.RTS {
		t.Error("RTS should be true")
	}

	if !st.mode.InitialStatusBits.DTR {
		t.Error("DTR should be true")
	}
}

func TestSerialTransport_Constants(t *testing.T) {
	if TransportCRC != 0x01 {
		t.Errorf("Expected TransportCRC to be 0x01, got 0x%02x", TransportCRC)
	}

	if MaxMsgSize != 4*1024 {
		t.Errorf("Expected MaxMsgSize to be 4096, got %d", MaxMsgSize)
	}
}
