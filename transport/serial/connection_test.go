package serial

import (
	"context"
	"kinetica-protocol/protocol/message"
	"testing"
	"time"
)

func TestConnection_Constants(t *testing.T) {
	if TransportCRC != 0x01 {
		t.Errorf("Expected TransportCRC to be 0x01, got 0x%02x", TransportCRC)
	}
	
	if MaxMsgSize != 4*1024 {
		t.Errorf("Expected MaxMsgSize to be 4096, got %d", MaxMsgSize)
	}
}

func TestConnection_State_ValidateInterface(t *testing.T) {
	ctx := context.Background()
	
	conn := &Connection{
		ctx:            ctx,
		readTimeout:    5 * time.Second,
		transportCRC:   message.TransportCRC8,
		maxMessageSize: 1024,
	}
	
	if conn.readTimeout != 5*time.Second {
		t.Errorf("Expected readTimeout 5s, got %v", conn.readTimeout)
	}
	
	if conn.transportCRC != message.TransportCRC8 {
		t.Errorf("Expected CRC8, got %v", conn.transportCRC)
	}
	
	if conn.maxMessageSize != 1024 {
		t.Errorf("Expected maxMessageSize 1024, got %d", conn.maxMessageSize)
	}
}

func TestConnection_PacketIDGeneration(t *testing.T) {
	ctx := context.Background()
	
	conn := &Connection{
		ctx: ctx,
	}
	
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
	ctx := context.Background()
	
	conn := &Connection{
		ctx: ctx,
	}
	
	conn.packetID.Store(255)
	
	id := conn.getNextPacketID()
	if id != 0 {
		t.Errorf("Expected ID to wrap to 0, got %d", id)
	}
}