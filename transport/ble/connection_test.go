package ble

import (
	"context"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"testing"
	"time"
)

func TestBleReader_Read(t *testing.T) {
	rxBuffer := make(chan []byte, 1)
	reader := &bleReader{
		rxBuffer:    rxBuffer,
		readTimeout: 100 * time.Millisecond,
	}

	testData := []byte("hello")
	rxBuffer <- testData

	buf := make([]byte, 10)
	n, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes read, got %d", n)
	}
	if string(buf[:n]) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(buf[:n]))
	}
}

func TestBleReader_ReadTimeout(t *testing.T) {
	rxBuffer := make(chan []byte)
	reader := &bleReader{
		rxBuffer:    rxBuffer,
		readTimeout: 50 * time.Millisecond,
	}

	buf := make([]byte, 10)
	start := time.Now()
	n, err := reader.Read(buf)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected timeout error")
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes read, got %d", n)
	}
	if elapsed < 40*time.Millisecond || elapsed > 70*time.Millisecond {
		t.Errorf("Expected timeout around 50ms, got %v", elapsed)
	}
}

func TestBleReader_ReadEOF(t *testing.T) {
	rxBuffer := make(chan []byte)
	reader := &bleReader{
		rxBuffer:    rxBuffer,
		readTimeout: 100 * time.Millisecond,
	}

	close(rxBuffer)

	buf := make([]byte, 10)
	n, err := reader.Read(buf)
	if err == nil {
		t.Fatal("Expected EOF error")
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes read, got %d", n)
	}
}

func TestBleReader_PartialRead(t *testing.T) {
	rxBuffer := make(chan []byte, 1)
	reader := &bleReader{
		rxBuffer:    rxBuffer,
		readTimeout: 100 * time.Millisecond,
	}

	testData := []byte("hello world")
	rxBuffer <- testData

	buf := make([]byte, 5)
	n, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("First read error = %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(buf) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(buf))
	}

	buf2 := make([]byte, 10)
	n, err = reader.Read(buf2)
	if err != nil {
		t.Fatalf("Second read error = %v", err)
	}
	if n != 6 {
		t.Errorf("Expected 6 bytes, got %d", n)
	}
	if string(buf2[:n]) != " world" {
		t.Errorf("Expected ' world', got '%s'", string(buf2[:n]))
	}
}

func TestConnection_State(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := &Connection{
		ctx: ctx,
	}

	state := conn.State()
	if state != transport.StateConnected {
		t.Errorf("Expected StateConnected, got %v", state)
	}

	cancel()

	state = conn.State()
	if state != transport.StateDisconnected {
		t.Errorf("Expected StateDisconnected after cancellation, got %v", state)
	}
}

func TestConnection_getNextPacketID(t *testing.T) {
	conn := &Connection{}

	id1 := conn.getNextPacketID()
	id2 := conn.getNextPacketID()

	if id1 == id2 {
		t.Error("Expected different packet IDs")
	}

	if id1 != 1 {
		t.Errorf("Expected first ID to be 1, got %d", id1)
	}

	if id2 != 2 {
		t.Errorf("Expected second ID to be 2, got %d", id2)
	}
}

func TestConnection_PacketID_Wraparound(t *testing.T) {
	conn := &Connection{}

	conn.packetID.Store(255)

	id := conn.getNextPacketID()
	if id != 0 {
		t.Errorf("Expected ID to wrap to 0, got %d", id)
	}
}

func TestConnection_Send_NilMessage(t *testing.T) {
	ctx := context.Background()
	conn := &Connection{
		ctx: ctx,
	}

	err := conn.Send(nil, message.MsgTypeHeartbeat)
	if err == nil {
		t.Fatal("Expected error for nil message")
	}

	if err.Error() != "invalid message size: message is nil" {
		t.Errorf("Expected invalid message size error, got %v", err)
	}
}

func TestConnection_Send_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	conn := &Connection{
		ctx: ctx,
	}

	heartbeat := &message.SensorHeartbeat{
		SensorID:  1,
		TimeStamp: 123456789,
		Battery:   85,
		Status:    message.Ok,
	}

	err := conn.Send(heartbeat, message.MsgTypeHeartbeat)
	if err == nil {
		t.Fatal("Expected error for cancelled context")
	}
}

func TestConstants(t *testing.T) {
	if TransportCRC != message.TransportCRC8 {
		t.Errorf("Expected TransportCRC to be CRC8, got %v", TransportCRC)
	}

	if MaxMessageSize != 255 {
		t.Errorf("Expected MaxMessageSize to be 255, got %d", MaxMessageSize)
	}
}
