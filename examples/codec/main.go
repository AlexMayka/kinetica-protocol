// Package main demonstrates basic usage of the Kinetica protocol codec.
// This example shows how to encode and decode different message types
// using the protocol's binary format with CRC8 transport layer.
package main

import (
	"fmt"
	"kinetica-protocol/protocol/codec"
	"kinetica-protocol/protocol/message"
)

func main() {
	// Simple heartbeat message
	heartbeat := &message.SensorHeartbeat{
		SensorID:  1,
		TimeStamp: 12345,
		Battery:   85,
		Status:    message.Ok,
	}
	testMessage(heartbeat, message.MsgTypeHeartbeat, "Heartbeat")

	// Sensor data message
	sensorData := &message.SensorData{
		SensorID:  1,
		TimeStamp: 12345,
		Data: message.Data{
			Type:   message.Accelerometer,
			Values: []float32{1.2, 3.4, 5.6},
		},
	}
	testMessage(sensorData, message.MsgTypeSensorData, "SensorData")

	// Multi-sensor data (new feature)
	sensorDataMulti := &message.SensorDataMulti{
		SensorID:  1,
		TimeStamp: 12345,
		Data: []message.Data{
			{Type: message.Accelerometer, Values: []float32{1.2, -0.5, 9.8}},
			{Type: message.Gyroscope, Values: []float32{0.1, 0.2, 0.3}},
		},
	}
	testMessage(sensorDataMulti, message.MsgTypeSensorDataMulti, "SensorDataMulti")

	// Relayed message (for ESP-NOW mesh)
	originalData, _ := codec.Marshal(heartbeat, 1, message.MsgTypeHeartbeat, message.TransportNone)
	relayedMessage := &message.RelayedMessage{
		RelayID:      10,
		OriginalData: originalData,
	}
	testMessage(relayedMessage, message.MsgTypeRelayed, "RelayedMessage")
}

// testMessage demonstrates encoding and decoding a message with the Kinetica protocol.
// It takes a message, encodes it to binary format, then decodes it back and compares results.
func testMessage(msg message.Message, msgType message.MsgType, name string) {
	fmt.Printf("=== %s ===\n", name)
	fmt.Printf("Original: %+v\n", msg)

	// Encode message
	data, err := codec.Marshal(msg, 1, msgType, message.TransportCRC8)
	if err != nil {
		fmt.Printf("Encode error: %v\n", err)
		return
	}
	fmt.Printf("Encoded (%d bytes): %x\n", len(data), data)

	// Decode message
	decoded, err := codec.Unmarshal(data, message.TransportCRC8)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}
	fmt.Printf("Decoded: %+v\n\n", decoded)
}
