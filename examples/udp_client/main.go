// Package main demonstrates a UDP client implementation using the Kinetica protocol.
// This example shows UDP datagram-based communication for sensor data transmission.
package main

import (
	"fmt"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport/net"
	"log"
	"time"
)

func main() {
	// Configure UDP connection
	config := net.Config{
		Address:      "localhost:8082",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	// Create UDP transport
	transport := net.NewUDP(config)
	defer transport.Close()

	// Connect to server
	conn, err := transport.Connection()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to UDP server")

	// Register sensor
	registration := &message.Registration{
		SensorID:     2,
		DeviceType:   message.DeviceType6Axis,
		Capabilities: message.CapAccelerometer,
		FWVersion:    0x0101,
	}

	if err := conn.Send(registration, message.MsgTypeRegister); err != nil {
		log.Fatalf("Failed to send registration: %v", err)
	}
	fmt.Println("Registration sent")

	// Send periodic data
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Send sensor data
		sensorData := &message.SensorData{
			SensorID:  2,
			TimeStamp: uint32(time.Now().Unix()),
			Data: message.Data{
				Type:   message.Accelerometer,
				Values: []float32{0.1, 0.2, 9.8},
			},
		}

		if err := conn.Send(sensorData, message.MsgTypeSensorData); err != nil {
			fmt.Printf("Send error: %v\n", err)
			break
		}
		fmt.Printf("Sent UDP data: %v\n", sensorData.Data.Values)
	}
}