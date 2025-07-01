package udp

import (
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"testing"
	"time"
)

func TestTransportUDP_Basic(t *testing.T) {
	config := &transport.Config{
		Type:           transport.UDP,
		Address:        ":0", // Автоматический выбор порта
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    10 * time.Second,
		DefaultCRCType: message.TransportNone,
	}

	udpTransport := NewTransportUDP(config)
	defer udpTransport.Close()

	// Проверяем тип
	if udpTransport.Type() != transport.UDP {
		t.Errorf("Expected UDP type, got %v", udpTransport.Type())
	}

	// Тестируем Listen
	connections, err := udpTransport.Listen()
	if err != nil {
		t.Fatalf("Failed to start listening: %v", err)
	}

	// Должны получить одно соединение
	select {
	case conn := <-connections:
		if conn == nil {
			t.Fatal("Received nil connection")
		}
		t.Logf("UDP server listening on %s", conn.RemoteAddr())
		conn.Close()
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for UDP connection")
	}

	// Канал должен закрыться при закрытии транспорта
	udpTransport.Close()
	
	select {
	case _, ok := <-connections:
		if ok {
			t.Error("Channel should be closed")
		}
	case <-time.After(1 * time.Second):
		t.Error("Channel should close immediately")
	}
}

func TestTransportUDP_ConvertAddr(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		expectError bool
	}{
		{"valid address", "localhost:8080", false},
		{"valid IP", "127.0.0.1:8080", false},
		{"empty address", "", true},
		{"invalid format", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &transport.Config{
				Type:    transport.UDP,
				Address: tt.address,
			}

			udpTransport := NewTransportUDP(config).(*TransportUDP)
			
			addr, err := udpTransport.ConvertAddr()
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if addr == nil {
					t.Error("Expected address, got nil")
				}
			}
		})
	}
}