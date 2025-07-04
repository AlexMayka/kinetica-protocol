package codec

import (
	"errors"
	"kinetica-protocol/protocol/message"
	"testing"
)

func TestMarshal_Registration(t *testing.T) {
	msg := &message.Registration{
		SensorID:     1,
		DeviceType:   message.DeviceType9Axis,
		Capabilities: message.CapAccelerometer,
		FWVersion:    0x0100,
	}

	data, err := Marshal(msg, 1, message.MsgTypeRegister, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshal returned empty data")
	}

	if data[0] != 'K' || data[1] != 'N' {
		t.Errorf("Expected magic bytes 'KN', got %c%c", data[0], data[1])
	}
}

func TestMarshal_SensorData(t *testing.T) {
	msg := &message.SensorData{
		SensorID:  1,
		TimeStamp: 12345,
		Data: message.Data{
			Type:   message.Accelerometer,
			Values: []float32{1.2, -0.5, 9.8},
		},
	}

	data, err := Marshal(msg, 1, message.MsgTypeSensorData, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshal returned empty data")
	}

	if data[0] != 'K' || data[1] != 'N' {
		t.Errorf("Expected magic bytes 'KN', got %c%c", data[0], data[1])
	}
}

func TestMarshal_SensorHeartbeat(t *testing.T) {
	msg := &message.SensorHeartbeat{
		SensorID:  1,
		TimeStamp: 12345,
		Battery:   85,
		Status:    message.Ok,
	}

	data, err := Marshal(msg, 1, message.MsgTypeHeartbeat, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshal returned empty data")
	}

	if data[0] != 'K' || data[1] != 'N' {
		t.Errorf("Expected magic bytes 'KN', got %c%c", data[0], data[1])
	}
}

func TestMarshal_Ack(t *testing.T) {
	msg := &message.Ack{
		SensorID:  1,
		MessageID: 1,
		Status:    message.AckOK,
	}

	data, err := Marshal(msg, 1, message.MsgTypeAck, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshal returned empty data")
	}

	if data[0] != 'K' || data[1] != 'N' {
		t.Errorf("Expected magic bytes 'KN', got %c%c", data[0], data[1])
	}
}

func TestMarshal_CustomData(t *testing.T) {
	msg := &message.CustomData{
		SensorID:  1,
		TimeStamp: 12345,
		DataType:  message.CustomTypeBinary,
		Data: []message.Item{
			{
				Key:    message.ConfigKeySampleRate,
				Length: 4,
				Value:  []byte{0x01, 0x02, 0x03, 0x04},
			},
		},
	}

	data, err := Marshal(msg, 1, message.MsgTypeCustom, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshal returned empty data")
	}

	if data[0] != 'K' || data[1] != 'N' {
		t.Errorf("Expected magic bytes 'KN', got %c%c", data[0], data[1])
	}
}

func TestMarshal_RelayedMessage(t *testing.T) {
	originalMsg := &message.SensorData{
		SensorID:  1,
		TimeStamp: 12345,
		Data: message.Data{
			Type:   message.Accelerometer,
			Values: []float32{1.0, 2.0, 3.0},
		},
	}

	originalData, err := Marshal(originalMsg, 1, message.MsgTypeSensorData, message.TransportNone)
	if err != nil {
		t.Fatalf("Failed to marshal original message: %v", err)
	}

	// Now create relayed message
	msg := &message.RelayedMessage{
		RelayID:      10,
		OriginalData: originalData,
	}

	data, err := Marshal(msg, 2, message.MsgTypeRelayed, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshal returned empty data")
	}

	if data[0] != 'K' || data[1] != 'N' {
		t.Errorf("Expected magic bytes 'KN', got %c%c", data[0], data[1])
	}

	// Size should be header(6) + relayID(1) + dataLength(2) + originalData(25)
	expectedSize := 6 + 1 + 2 + len(originalData)
	if len(data) != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, len(data))
	}
}

func TestMarshal_SensorDataMulti(t *testing.T) {
	msg := &message.SensorDataMulti{
		SensorID:  1,
		TimeStamp: 12345,
		Data: []message.Data{
			{Type: message.Accelerometer, Values: []float32{1.0, 2.0, 3.0}},
			{Type: message.Gyroscope, Values: []float32{4.0, 5.0, 6.0}},
			{Type: message.Quaternion, Values: []float32{0.0, 0.0, 0.0, 1.0}},
		},
	}

	data, err := Marshal(msg, 1, message.MsgTypeSensorDataMulti, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshal returned empty data")
	}

	if data[0] != 'K' || data[1] != 'N' {
		t.Errorf("Expected magic bytes 'KN', got %c%c", data[0], data[1])
	}

	expectedSize := 6 + 1 + 4 + 1 + (1 + 1 + 12) + (1 + 1 + 12) + (1 + 1 + 16)
	if len(data) != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, len(data))
	}
}

func TestMarshal_AllCRCTypes(t *testing.T) {
	msg := &message.SensorHeartbeat{
		SensorID:  1,
		TimeStamp: 12345,
		Battery:   85,
		Status:    message.Ok,
	}

	crcTypes := []message.TransportCRC{
		message.TransportNone,
		message.TransportCRC8,
		message.TransportCRC16,
		message.TransportCRC32,
		message.TransportLength,
	}

	for _, crcType := range crcTypes {
		t.Run(string(rune(crcType)), func(t *testing.T) {
			data, err := Marshal(msg, 1, message.MsgTypeHeartbeat, crcType)
			if err != nil {
				t.Fatalf("Marshal failed for CRC type %v: %v", crcType, err)
			}

			if len(data) == 0 {
				t.Fatal("Marshal returned empty data")
			}

			if data[0] != 'K' || data[1] != 'N' {
				t.Errorf("Expected magic bytes 'KN', got %c%c", data[0], data[1])
			}
		})
	}
}

func TestUnmarshal_Registration(t *testing.T) {
	originalMsg := &message.Registration{
		SensorID:     1,
		DeviceType:   message.DeviceType9Axis,
		Capabilities: message.CapAccelerometer | message.CapGyroscope,
		FWVersion:    0x0100,
	}

	data, err := Marshal(originalMsg, 1, message.MsgTypeRegister, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	msg, err := Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	registration, ok := msg.(*message.Registration)
	if !ok {
		t.Fatalf("Expected Registration message, got %T", msg)
	}

	if registration.SensorID != originalMsg.SensorID {
		t.Errorf("SensorID mismatch: expected %d, got %d", originalMsg.SensorID, registration.SensorID)
	}
	if registration.DeviceType != originalMsg.DeviceType {
		t.Errorf("DeviceType mismatch: expected %d, got %d", originalMsg.DeviceType, registration.DeviceType)
	}
	if registration.Capabilities != originalMsg.Capabilities {
		t.Errorf("Capabilities mismatch: expected %d, got %d", originalMsg.Capabilities, registration.Capabilities)
	}
	if registration.FWVersion != originalMsg.FWVersion {
		t.Errorf("FWVersion mismatch: expected %d, got %d", originalMsg.FWVersion, registration.FWVersion)
	}
}

func TestUnmarshal_SensorData(t *testing.T) {
	originalMsg := &message.SensorData{
		SensorID:  1,
		TimeStamp: 12345,
		Data: message.Data{
			Type:   message.Accelerometer,
			Values: []float32{1.2, -0.5, 9.8},
		},
	}

	data, err := Marshal(originalMsg, 1, message.MsgTypeSensorData, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	msg, err := Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	sensorData, ok := msg.(*message.SensorData)
	if !ok {
		t.Fatalf("Expected SensorData message, got %T", msg)
	}

	if sensorData.SensorID != originalMsg.SensorID {
		t.Errorf("SensorID mismatch: expected %d, got %d", originalMsg.SensorID, sensorData.SensorID)
	}
	if sensorData.TimeStamp != originalMsg.TimeStamp {
		t.Errorf("TimeStamp mismatch: expected %d, got %d", originalMsg.TimeStamp, sensorData.TimeStamp)
	}
	if sensorData.Data.Type != originalMsg.Data.Type {
		t.Errorf("Type mismatch: expected %d, got %d", originalMsg.Data.Type, sensorData.Data.Type)
	}
	if len(sensorData.Data.Values) != len(originalMsg.Data.Values) {
		t.Errorf("Values length mismatch: expected %d, got %d", len(originalMsg.Data.Values), len(sensorData.Data.Values))
	}
}

func TestUnmarshal_SensorHeartbeat(t *testing.T) {
	originalMsg := &message.SensorHeartbeat{
		SensorID:  1,
		TimeStamp: 12345,
		Battery:   85,
		Status:    message.Ok,
	}

	data, err := Marshal(originalMsg, 1, message.MsgTypeHeartbeat, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	msg, err := Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	heartbeat, ok := msg.(*message.SensorHeartbeat)
	if !ok {
		t.Fatalf("Expected SensorHeartbeat message, got %T", msg)
	}

	if heartbeat.SensorID != originalMsg.SensorID {
		t.Errorf("SensorID mismatch: expected %d, got %d", originalMsg.SensorID, heartbeat.SensorID)
	}
	if heartbeat.TimeStamp != originalMsg.TimeStamp {
		t.Errorf("TimeStamp mismatch: expected %d, got %d", originalMsg.TimeStamp, heartbeat.TimeStamp)
	}
	if heartbeat.Battery != originalMsg.Battery {
		t.Errorf("Battery mismatch: expected %d, got %d", originalMsg.Battery, heartbeat.Battery)
	}
	if heartbeat.Status != originalMsg.Status {
		t.Errorf("Status mismatch: expected %d, got %d", originalMsg.Status, heartbeat.Status)
	}
}

func TestUnmarshal_Ack(t *testing.T) {
	originalMsg := &message.Ack{
		SensorID:  1,
		MessageID: 1,
		Status:    message.AckOK,
	}

	data, err := Marshal(originalMsg, 1, message.MsgTypeAck, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	msg, err := Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	ack, ok := msg.(*message.Ack)
	if !ok {
		t.Fatalf("Expected Ack message, got %T", msg)
	}

	if ack.SensorID != originalMsg.SensorID {
		t.Errorf("SensorID mismatch: expected %d, got %d", originalMsg.SensorID, ack.SensorID)
	}
	if ack.MessageID != originalMsg.MessageID {
		t.Errorf("MessageID mismatch: expected %d, got %d", originalMsg.MessageID, ack.MessageID)
	}
	if ack.Status != originalMsg.Status {
		t.Errorf("Status mismatch: expected %d, got %d", originalMsg.Status, ack.Status)
	}
}

func TestUnmarshal_RelayedMessage(t *testing.T) {
	originalMsg := &message.SensorHeartbeat{
		SensorID:  5,
		TimeStamp: 54321,
		Battery:   75,
		Status:    message.Ok,
	}

	originalData, err := Marshal(originalMsg, 3, message.MsgTypeHeartbeat, message.TransportNone)
	if err != nil {
		t.Fatalf("Failed to marshal original message: %v", err)
	}

	relayedMsg := &message.RelayedMessage{
		RelayID:      20,
		OriginalData: originalData,
	}

	data, err := Marshal(relayedMsg, 4, message.MsgTypeRelayed, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	msg, err := Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	relayed, ok := msg.(*message.RelayedMessage)
	if !ok {
		t.Fatalf("Expected RelayedMessage, got %T", msg)
	}

	if relayed.RelayID != relayedMsg.RelayID {
		t.Errorf("RelayID mismatch: expected %d, got %d", relayedMsg.RelayID, relayed.RelayID)
	}

	if len(relayed.OriginalData) != len(originalData) {
		t.Errorf("OriginalData length mismatch: expected %d, got %d", len(originalData), len(relayed.OriginalData))
	}

	innerMsg, err := Unmarshal(relayed.OriginalData, message.TransportNone)
	if err != nil {
		t.Fatalf("Failed to unmarshal inner message: %v", err)
	}

	heartbeat, ok := innerMsg.(*message.SensorHeartbeat)
	if !ok {
		t.Fatalf("Expected inner message to be SensorHeartbeat, got %T", innerMsg)
	}

	if heartbeat.SensorID != originalMsg.SensorID || heartbeat.Battery != originalMsg.Battery {
		t.Error("Inner message data mismatch after relay")
	}
}

func TestUnmarshal_SensorDataMulti(t *testing.T) {
	originalMsg := &message.SensorDataMulti{
		SensorID:  1,
		TimeStamp: 12345,
		Data: []message.Data{
			{Type: message.Accelerometer, Values: []float32{1.0, 2.0, 3.0}},
			{Type: message.Gyroscope, Values: []float32{4.0, 5.0, 6.0}},
		},
	}

	data, err := Marshal(originalMsg, 1, message.MsgTypeSensorDataMulti, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	msg, err := Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	sensorDataMulti, ok := msg.(*message.SensorDataMulti)
	if !ok {
		t.Fatalf("Expected SensorDataMulti, got %T", msg)
	}

	if sensorDataMulti.SensorID != originalMsg.SensorID {
		t.Errorf("SensorID mismatch: expected %d, got %d", originalMsg.SensorID, sensorDataMulti.SensorID)
	}
	if sensorDataMulti.TimeStamp != originalMsg.TimeStamp {
		t.Errorf("TimeStamp mismatch: expected %d, got %d", originalMsg.TimeStamp, sensorDataMulti.TimeStamp)
	}
	if len(sensorDataMulti.Data) != len(originalMsg.Data) {
		t.Errorf("Data length mismatch: expected %d, got %d", len(originalMsg.Data), len(sensorDataMulti.Data))
	}

	for i, data := range sensorDataMulti.Data {
		if data.Type != originalMsg.Data[i].Type {
			t.Errorf("Data[%d].Type mismatch: expected %d, got %d", i, originalMsg.Data[i].Type, data.Type)
		}
		if len(data.Values) != len(originalMsg.Data[i].Values) {
			t.Errorf("Data[%d].Values length mismatch: expected %d, got %d", i, len(originalMsg.Data[i].Values), len(data.Values))
		}
	}
}

func TestUnmarshal_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		transport   message.TransportCRC
		expectedErr error
	}{
		{
			name:        "message too short",
			data:        []byte{0x4B, 0x4E},
			transport:   message.TransportNone,
			expectedErr: ErrMessageTooShort,
		},
		{
			name:        "invalid magic bytes",
			data:        []byte{0x00, 0x00, 0x01, 0x01, 0x01, 0x00},
			transport:   message.TransportNone,
			expectedErr: ErrInvalidMagicBytes,
		},
		{
			name:        "empty data",
			data:        []byte{},
			transport:   message.TransportNone,
			expectedErr: ErrMessageTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Unmarshal(tt.data, tt.transport)
			if err == nil {
				t.Error("Expected error, got nil")
				return
			}

			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestMarshal_Unmarshal_RoundTrip(t *testing.T) {
	messages := []struct {
		name    string
		msg     message.Message
		msgType message.MsgType
	}{
		{
			name: "Registration",
			msg: &message.Registration{
				SensorID:     42,
				DeviceType:   message.DeviceType9Axis,
				Capabilities: message.CapAccelerometer | message.CapGyroscope,
				FWVersion:    0x0100,
			},
			msgType: message.MsgTypeRegister,
		},
		{
			name: "SensorData",
			msg: &message.SensorData{
				SensorID:  42,
				TimeStamp: 12345,
				Data: message.Data{
					Type:   message.Accelerometer,
					Values: []float32{1.0, 2.0, 3.0},
				},
			},
			msgType: message.MsgTypeSensorData,
		},
		{
			name: "SensorHeartbeat",
			msg: &message.SensorHeartbeat{
				SensorID:  42,
				TimeStamp: 12345,
				Battery:   85,
				Status:    message.Ok,
			},
			msgType: message.MsgTypeHeartbeat,
		},
		{
			name: "Ack",
			msg: &message.Ack{
				SensorID:  42,
				MessageID: 1,
				Status:    message.AckOK,
			},
			msgType: message.MsgTypeAck,
		},
		{
			name: "SensorDataMulti",
			msg: &message.SensorDataMulti{
				SensorID:  42,
				TimeStamp: 12345,
				Data: []message.Data{
					{Type: message.Accelerometer, Values: []float32{1.0, 2.0, 3.0}},
					{Type: message.Gyroscope, Values: []float32{4.0, 5.0, 6.0}},
				},
			},
			msgType: message.MsgTypeSensorDataMulti,
		},
		{
			name: "RelayedMessage",
			msg: &message.RelayedMessage{
				RelayID:      10,
				OriginalData: []byte{0x4B, 0x4E, 0x01, 0x04, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05},
			},
			msgType: message.MsgTypeRelayed,
		},
	}

	transports := []message.TransportCRC{
		message.TransportNone,
		message.TransportCRC8,
		message.TransportCRC16,
		message.TransportCRC32,
		message.TransportLength,
	}

	for _, msg := range messages {
		for _, transport := range transports {
			t.Run(msg.name+"_"+string(rune(transport)), func(t *testing.T) {
				data, err := Marshal(msg.msg, 1, msg.msgType, transport)
				if err != nil {
					t.Fatalf("Marshal failed: %v", err)
				}

				unmarshaled, err := Unmarshal(data, transport)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}

				if unmarshaled == nil {
					t.Fatal("Unmarshal returned nil message")
				}
			})
		}
	}
}
