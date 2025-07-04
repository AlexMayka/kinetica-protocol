package net

import (
	"testing"
	"time"
)

func TestUDPTransport_NewUDP(t *testing.T) {
	config := Config{
		Address:      "localhost:8080",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	udpTransport := NewUDP(config)
	
	if udpTransport == nil {
		t.Fatal("Expected non-nil transport")
	}
	
	if udpTransport.config.Address != "localhost:8080" {
		t.Errorf("Expected address localhost:8080, got %s", udpTransport.config.Address)
	}
}

func TestUDPTransport_Connection_InvalidAddress(t *testing.T) {
	config := Config{
		Address: "invalid:address:format",
	}

	udpTransport := NewUDP(config)
	
	_, err := udpTransport.Connection()
	if err == nil {
		t.Fatal("Expected error for invalid address")
	}
}

func TestUDPTransport_Connection_Success(t *testing.T) {
	config := Config{
		Address: "127.0.0.1:8888",
	}

	udpTransport := NewUDP(config)
	
	conn, err := udpTransport.Connection()
	if err != nil {
		t.Fatalf("Connection() error = %v", err)
	}
	
	if conn == nil {
		t.Fatal("Expected non-nil connection")
	}
	
	conn.Close()
	udpTransport.Close()
}

func TestUDPTransport_Listen(t *testing.T) {
	config := Config{
		Address: "localhost:0",
	}

	udpTransport := NewUDP(config)
	
	ch, err := udpTransport.Listen()
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	
	if ch == nil {
		t.Fatal("Expected non-nil channel")
	}
	
	select {
	case conn := <-ch:
		if conn == nil {
			t.Fatal("Expected non-nil connection")
		}
		conn.Close()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected connection from channel")
	}
	
	udpTransport.Close()
}

func TestUDPTransport_Listen_InvalidAddress(t *testing.T) {
	config := Config{
		Address: "invalid:address:format",
	}

	udpTransport := NewUDP(config)
	
	_, err := udpTransport.Listen()
	if err == nil {
		t.Fatal("Expected error for invalid address")
	}
}

func TestUDPTransport_Close(t *testing.T) {
	config := Config{
		Address: "localhost:0",
	}

	udpTransport := NewUDP(config)
	
	err := udpTransport.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
	
	select {
	case <-udpTransport.ctx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Error("Context not cancelled after Close()")
	}
}

func TestUDPTransport_Constants(t *testing.T) {
	if UDPTransportCRC != 0x01 {
		t.Errorf("Expected UDPTransportCRC to be 0x01, got 0x%02x", UDPTransportCRC)
	}
	
	if UDPMaxMessageSize != 1472 {
		t.Errorf("Expected UDPMaxMessageSize to be 1472, got %d", UDPMaxMessageSize)
	}
}

func TestUDPTransport_AddressResolution(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid localhost",
			address: "localhost:8080",
			wantErr: false,
		},
		{
			name:    "valid IP",
			address: "127.0.0.1:9090",
			wantErr: false,
		},
		{
			name:    "zero port",
			address: ":8889",
			wantErr: false,
		},
		{
			name:    "invalid format",
			address: "invalid:address:format",
			wantErr: true,
		},
		{
			name:    "missing port",
			address: "localhost",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Address: tt.address}
			udpTransport := NewUDP(config)
			
			_, err := udpTransport.Connection()
			if (err != nil) != tt.wantErr {
				t.Errorf("Connection() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if err == nil {
				udpTransport.Close()
			}
		})
	}
}