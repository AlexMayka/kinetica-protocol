package tcp

import (
	"errors"
	"kinetica-protocol/transport"
	"net"
	"testing"
	"time"
)

func TestTransportTCP_Type(t *testing.T) {
	config := &transport.Config{Address: "localhost:8080"}
	tp := NewTransportTCP(config)

	if tp.Type() != transport.TCP {
		t.Errorf("Expected type %v, got %v", transport.TCP, tp.Type())
	}
}

func TestTransportTCP_ConvertAddr(t *testing.T) {
	tests := []struct {
		name    string
		config  *transport.Config
		wantErr bool
		errType error
	}{
		{
			name:    "valid address",
			config:  &transport.Config{Address: "localhost:8080"},
			wantErr: false,
		},
		{
			name:    "valid IP address",
			config:  &transport.Config{Address: "127.0.0.1:9090"},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errType: transport.ErrInvalidConfig,
		},
		{
			name:    "empty address",
			config:  &transport.Config{Address: ""},
			wantErr: true,
			errType: transport.ErrInvalidAddrTrans,
		},
		{
			name:    "invalid address format",
			config:  &transport.Config{Address: "invalid-address"},
			wantErr: true,
			errType: transport.ErrInvalidAddrTrans,
		},
		{
			name:    "invalid port",
			config:  &transport.Config{Address: "localhost:99999"},
			wantErr: true,
			errType: transport.ErrInvalidAddrTrans,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := NewTransportTCP(tt.config)
			addr, err := tp.ConvertAddr()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("Expected error type %v, got %v", tt.errType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if addr == nil {
				t.Error("Expected valid address, got nil")
			}
		})
	}
}

func TestTransportTCP_Connection(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer func(listener net.Listener) { _ = listener.Close() }(listener)

	serverAddr := listener.Addr().String()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	tests := []struct {
		name    string
		address string
		wantErr bool
		errType error
	}{
		{
			name:    "successful connection",
			address: serverAddr,
			wantErr: false,
		},
		{
			name:    "connection refused",
			address: "localhost:0",
			wantErr: true,
			errType: transport.ErrConnectionFailed,
		},
		{
			name:    "invalid address",
			address: "invalid-host:8080",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &transport.Config{
				Address:      tt.address,
				WriteTimeout: time.Second,
				ReadTimeout:  time.Second,
			}
			tp := NewTransportTCP(config)

			conn, err := tp.Connection()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					if conn != nil {
						_ = conn.Close()
					}
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("Expected error type %v, got %v", tt.errType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if conn == nil {
				t.Error("Expected connection, got nil")
				return
			}

			_ = conn.Close()
		})
	}
}

func TestTransportTCP_Listen(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
		errType error
	}{
		{
			name:    "valid listen address",
			address: "localhost:0",
			wantErr: false,
		},
		{
			name:    "invalid address",
			address: "invalid-host:8080",
			wantErr: true,
			errType: transport.ErrInvalidAddrTrans,
		},
		{
			name:    "empty address",
			address: "",
			wantErr: true,
			errType: transport.ErrInvalidAddrTrans,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &transport.Config{
				Address:      tt.address,
				WriteTimeout: time.Second,
				ReadTimeout:  time.Second,
			}
			tp := NewTransportTCP(config)

			ch, err := tp.Listen()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					if ch != nil {
						_ = tp.Close()
					}
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("Expected error type %v, got %v", tt.errType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if ch == nil {
				t.Error("Expected channel, got nil")
				return
			}

			go func() {
				addr := tp.listener.Addr().String()
				conn, err := net.Dial("tcp", addr)
				if err == nil {
					_ = conn.Close()
				}
			}()

			select {
			case conn := <-ch:
				if conn == nil {
					t.Error("Received nil connection")
				} else {
					_ = conn.Close()
				}
			case <-time.After(time.Second):
				t.Error("Timeout waiting for connection")
			}

			_ = tp.Close()
		})
	}
}

func TestTransportTCP_Close(t *testing.T) {
	config := &transport.Config{Address: "localhost:0"}
	tp := NewTransportTCP(config)

	err := tp.Close()
	if err != nil {
		t.Errorf("Unexpected error closing transport without listener: %v", err)
	}

	_, err = tp.Listen()
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}

	err = tp.Close()
	if err != nil {
		t.Errorf("Unexpected error closing transport with listener: %v", err)
	}

	select {
	case <-tp.ctx.Done():
	default:
		t.Error("Expected context to be canceled")
	}
}

func TestTransportTCP_ContextCancellation(t *testing.T) {
	config := &transport.Config{Address: "localhost:0"}
	tp := NewTransportTCP(config)

	ch, err := tp.Listen()
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}

	_ = tp.Close()

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for channel to close")
	}
}
