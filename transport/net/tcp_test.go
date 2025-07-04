package net

import (
	"strings"
	"testing"
	"time"
)

func TestTCPTransport_NewTCP(t *testing.T) {
	config := Config{
		Address:      "localhost:8080",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	tcpTransport := NewTCP(config)
	
	if tcpTransport == nil {
		t.Fatal("Expected non-nil transport")
	}
	
	if tcpTransport.config.Address != "localhost:8080" {
		t.Errorf("Expected address localhost:8080, got %s", tcpTransport.config.Address)
	}
}

func TestTCPTransport_Connection_InvalidAddress(t *testing.T) {
	config := Config{
		Address: "invalid:address:format",
	}

	tcpTransport := NewTCP(config)
	
	_, err := tcpTransport.Connection()
	if err == nil {
		t.Fatal("Expected error for invalid address")
	}
}

func TestTCPTransport_Connection_ConnectionRefused(t *testing.T) {
	config := Config{
		Address: "localhost:9999",
	}

	tcpTransport := NewTCP(config)
	
	_, err := tcpTransport.Connection()
	if err == nil {
		t.Fatal("Expected connection refused error")
	}
	
	if !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("Expected connection error, got %v", err)
	}
}

func TestTCPTransport_Listen(t *testing.T) {
	config := Config{
		Address: "localhost:0",
	}

	tcpTransport := NewTCP(config)
	
	ch, err := tcpTransport.Listen()
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	
	if ch == nil {
		t.Fatal("Expected non-nil channel")
	}
	
	tcpTransport.Close()
}

func TestTCPTransport_Listen_InvalidAddress(t *testing.T) {
	config := Config{
		Address: "invalid:address:format",
	}

	tcpTransport := NewTCP(config)
	
	_, err := tcpTransport.Listen()
	if err == nil {
		t.Fatal("Expected error for invalid address")
	}
}

func TestTCPTransport_Close(t *testing.T) {
	config := Config{
		Address: "localhost:0",
	}

	tcpTransport := NewTCP(config)
	
	err := tcpTransport.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
	
	select {
	case <-tcpTransport.ctx.Done():
	case <-time.After(100 * time.Millisecond):
		t.Error("Context not cancelled after Close()")
	}
}

func TestTCPTransport_Constants(t *testing.T) {
	if TCPMaxMessageSize != 64*1024 {
		t.Errorf("Expected TCPMaxMessageSize to be 65536, got %d", TCPMaxMessageSize)
	}
}