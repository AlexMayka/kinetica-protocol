package net

import (
	"context"
	"fmt"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"net"
	"strings"
	"testing"
	"time"
)

type mockNetConn struct {
	writeData []byte
	readData  []byte
	readIndex int
	closed    bool
}

func (m *mockNetConn) Read(b []byte) (n int, err error) {
	if m.readIndex >= len(m.readData) {
		return 0, fmt.Errorf("EOF")
	}

	n = copy(b, m.readData[m.readIndex:])
	m.readIndex += n
	return n, nil
}

func (m *mockNetConn) Write(b []byte) (n int, err error) {
	if m.closed {
		return 0, fmt.Errorf("connection closed")
	}
	m.writeData = append(m.writeData, b...)
	return len(b), nil
}

func (m *mockNetConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockNetConn) LocalAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	return addr
}

func (m *mockNetConn) RemoteAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:8081")
	return addr
}

func (m *mockNetConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockNetConn) SetWriteDeadline(t time.Time) error { return nil }

func TestConnection_NewConnection(t *testing.T) {
	mock := &mockNetConn{}
	ctx := context.Background()

	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	if conn == nil {
		t.Fatal("Expected non-nil connection")
	}

	if conn.writeTimeout != 5*time.Second {
		t.Errorf("Expected writeTimeout 5s, got %v", conn.writeTimeout)
	}

	if conn.readTimeout != 10*time.Second {
		t.Errorf("Expected readTimeout 10s, got %v", conn.readTimeout)
	}

	if conn.transportCRC != message.TransportCRC8 {
		t.Errorf("Expected CRC8, got %v", conn.transportCRC)
	}

	if conn.maxMessageSize != 1024 {
		t.Errorf("Expected maxMessageSize 1024, got %d", conn.maxMessageSize)
	}
}

func TestConnection_Send(t *testing.T) {
	mock := &mockNetConn{}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	heartbeat := &message.SensorHeartbeat{
		SensorID:  1,
		TimeStamp: 123456789,
		Battery:   85,
		Status:    message.Ok,
	}

	err := conn.Send(heartbeat, message.MsgTypeHeartbeat)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if len(mock.writeData) == 0 {
		t.Fatal("Expected data to be written")
	}

	if mock.writeData[0] != 0x4B || mock.writeData[1] != 0x4E {
		t.Error("Expected Kinetica magic bytes")
	}
}

func TestConnection_Send_NilMessage(t *testing.T) {
	mock := &mockNetConn{}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	err := conn.Send(nil, message.MsgTypeHeartbeat)
	if err == nil {
		t.Fatal("Expected error for nil message")
	}

	if !strings.Contains(fmt.Sprintf("%v", err), "invalid message size") {
		t.Errorf("Expected invalid message size error, got %v", err)
	}
}

func TestConnection_Send_MessageTooLarge(t *testing.T) {
	mock := &mockNetConn{}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 10)

	heartbeat := &message.SensorHeartbeat{
		SensorID:  1,
		TimeStamp: 123456789,
		Battery:   85,
		Status:    message.Ok,
	}

	err := conn.Send(heartbeat, message.MsgTypeHeartbeat)
	if err == nil {
		t.Fatal("Expected error for large message")
	}

	if !strings.Contains(fmt.Sprintf("%v", err), "message too large") {
		t.Errorf("Expected message too large error, got %v", err)
	}
}

func TestConnection_Send_ContextCancelled(t *testing.T) {
	mock := &mockNetConn{}
	ctx, cancel := context.WithCancel(context.Background())
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	cancel()

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

	if !strings.Contains(fmt.Sprintf("%v", err), "context canceled") {
		t.Errorf("Expected context canceled error, got %v", err)
	}
}

func TestConnection_State(t *testing.T) {
	mock := &mockNetConn{
		readData: []byte{0x01, 0x02, 0x03},
	}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	state := conn.State()
	if state != transport.StateConnected {
		t.Errorf("Expected StateConnected, got %v", state)
	}
}

func TestConnection_State_Cancelled(t *testing.T) {
	mock := &mockNetConn{}
	ctx, cancel := context.WithCancel(context.Background())
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	cancel()

	state := conn.State()
	if state != transport.StateDisconnected {
		t.Errorf("Expected StateDisconnected, got %v", state)
	}
}

func TestConnection_RemoteAddr(t *testing.T) {
	mock := &mockNetConn{}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	addr := conn.RemoteAddr()
	if addr == nil {
		t.Fatal("Expected non-nil remote address")
	}

	if addr.String() != "127.0.0.1:8081" {
		t.Errorf("Expected remote address 127.0.0.1:8081, got %s", addr.String())
	}
}

func TestConnection_LocalAddr(t *testing.T) {
	mock := &mockNetConn{}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	addr := conn.LocalAddr()
	if addr == nil {
		t.Fatal("Expected non-nil local address")
	}

	if addr.String() != "127.0.0.1:8080" {
		t.Errorf("Expected local address 127.0.0.1:8080, got %s", addr.String())
	}
}

func TestConnection_Close(t *testing.T) {
	mock := &mockNetConn{}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	err := conn.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !mock.closed {
		t.Error("Expected mock connection to be closed")
	}
}

func TestConnection_getNextPacketID(t *testing.T) {
	mock := &mockNetConn{}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

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
	mock := &mockNetConn{}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	conn.packetID.Store(255)

	id := conn.getNextPacketID()
	if id != 0 {
		t.Errorf("Expected ID to wrap to 0, got %d", id)
	}
}

func TestConnection_Stats(t *testing.T) {
	mock := &mockNetConn{}
	ctx := context.Background()
	conn := NewConnection(mock, ctx, 5*time.Second, 10*time.Second, message.TransportCRC8, 1024)

	sentBytes, recvBytes, sentMsgs, recvMsgs := conn.Stats()

	if sentBytes != 0 || recvBytes != 0 || sentMsgs != 0 || recvMsgs != 0 {
		t.Error("Expected all stats to be zero initially")
	}
}
