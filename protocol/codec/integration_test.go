package codec_test

import (
	"kinetica-protocol/protocol/codec"
	"kinetica-protocol/protocol/message"
	"testing"
	"time"
)

func TestCodec_Integration_MarshalUnmarshal(t *testing.T) {
	registration := &message.Registration{
		SensorID:     1,
		DeviceType:   message.DeviceType9Axis,
		Capabilities: message.CapAccelerometer | message.CapGyroscope,
		FWVersion:    0x0100,
	}

	data, err := codec.Marshal(registration, 1, message.MsgTypeRegister, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal registration failed: %v", err)
	}
	t.Logf("Registration marshaled to %d bytes", len(data))

	unmarshaled, err := codec.Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal registration failed: %v", err)
	}

	reg, ok := unmarshaled.(*message.Registration)
	if !ok {
		t.Fatalf("Expected Registration, got %T", unmarshaled)
	}

	if reg.SensorID != registration.SensorID || reg.DeviceType != registration.DeviceType ||
		reg.Capabilities != registration.Capabilities || reg.FWVersion != registration.FWVersion {
		t.Error("Registration data mismatch after marshal/unmarshal")
	}

	sensorData := &message.SensorData{
		SensorID:  1,
		TimeStamp: uint32(time.Now().Unix()),
		Data: message.Data{
			Type:   message.Accelerometer,
			Values: []float32{1.2, -0.5, 9.8},
		},
	}

	data, err = codec.Marshal(sensorData, 1, message.MsgTypeSensorData, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal sensor data failed: %v", err)
	}
	t.Logf("Sensor data marshaled to %d bytes", len(data))

	unmarshaled, err = codec.Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal sensor data failed: %v", err)
	}

	sd, ok := unmarshaled.(*message.SensorData)
	if !ok {
		t.Fatalf("Expected SensorData, got %T", unmarshaled)
	}

	if sd.SensorID != sensorData.SensorID || sd.Data.Type != sensorData.Data.Type || len(sd.Data.Values) != len(sensorData.Data.Values) {
		t.Error("Sensor data mismatch after marshal/unmarshal")
	}

	heartbeat := &message.SensorHeartbeat{
		SensorID:  1,
		TimeStamp: uint32(time.Now().Unix()),
		Battery:   85,
		Status:    message.Ok,
	}

	data, err = codec.Marshal(heartbeat, 1, message.MsgTypeHeartbeat, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal heartbeat failed: %v", err)
	}
	t.Logf("Heartbeat marshaled to %d bytes", len(data))

	unmarshaled, err = codec.Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal heartbeat failed: %v", err)
	}

	hb, ok := unmarshaled.(*message.SensorHeartbeat)
	if !ok {
		t.Fatalf("Expected SensorHeartbeat, got %T", unmarshaled)
	}

	if hb.SensorID != heartbeat.SensorID || hb.Battery != heartbeat.Battery || hb.Status != heartbeat.Status {
		t.Error("Heartbeat data mismatch after marshal/unmarshal")
	}

	ack := &message.Ack{
		SensorID:  1,
		MessageID: 1,
		Status:    message.AckOK,
	}

	data, err = codec.Marshal(ack, 1, message.MsgTypeAck, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal ack failed: %v", err)
	}
	t.Logf("ACK marshaled to %d bytes", len(data))

	unmarshaled, err = codec.Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal ack failed: %v", err)
	}

	ackResult, ok := unmarshaled.(*message.Ack)
	if !ok {
		t.Fatalf("Expected Ack, got %T", unmarshaled)
	}

	if ackResult.SensorID != ack.SensorID || ackResult.MessageID != ack.MessageID || ackResult.Status != ack.Status {
		t.Error("ACK data mismatch after marshal/unmarshal")
	}

	// Test SensorDataMulti
	sensorDataMulti := &message.SensorDataMulti{
		SensorID:  1,
		TimeStamp: uint32(time.Now().Unix()),
		Data: []message.Data{
			{Type: message.Accelerometer, Values: []float32{1.2, -0.5, 9.8}},
			{Type: message.Gyroscope, Values: []float32{0.1, 0.2, 0.3}},
			{Type: message.Quaternion, Values: []float32{0.0, 0.0, 0.0, 1.0}},
		},
	}

	data, err = codec.Marshal(sensorDataMulti, 1, message.MsgTypeSensorDataMulti, message.TransportNone)
	if err != nil {
		t.Fatalf("Marshal sensor data multi failed: %v", err)
	}
	t.Logf("SensorDataMulti marshaled to %d bytes", len(data))

	unmarshaled, err = codec.Unmarshal(data, message.TransportNone)
	if err != nil {
		t.Fatalf("Unmarshal sensor data multi failed: %v", err)
	}

	sdm, ok := unmarshaled.(*message.SensorDataMulti)
	if !ok {
		t.Fatalf("Expected SensorDataMulti, got %T", unmarshaled)
	}

	if sdm.SensorID != sensorDataMulti.SensorID || len(sdm.Data) != len(sensorDataMulti.Data) {
		t.Error("SensorDataMulti data mismatch after marshal/unmarshal")
	}
}
