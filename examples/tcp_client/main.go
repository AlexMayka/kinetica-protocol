// Package main demonstrates a TCP client implementation using the Kinetica protocol.
// This example shows how to connect to a TCP server, register a sensor,
// and send periodic sensor data and heartbeat messages.
package main

import (
	"fmt"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport/net"
	"log"
	"time"
)

func main() {
	// Configure TCP connection
	config := net.Config{
		Address:      "localhost:8081",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	// Create TCP transport
	transport := net.NewTCP(config)
	defer transport.Close()

	// Connect to server
	conn, err := transport.Connection()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to server")

	// Register sensor
	registration := &message.Registration{
		SensorID:     1,
		DeviceType:   message.DeviceType9Axis,
		Capabilities: message.CapAccelerometer | message.CapGyroscope,
		FWVersion:    0x0100,
	}

	if err := conn.Send(registration, message.MsgTypeRegister); err != nil {
		log.Fatalf("Failed to send registration: %v", err)
	}
	fmt.Println("Registration sent")

	// Listen for server messages
	go func() {
		for {
			msg, err := conn.Receive()
			if err != nil {
				fmt.Printf("Receive error: %v\n", err)
				return
			}

			switch m := msg.(type) {
			case *message.Ack:
				fmt.Printf("Server ACK: status=%d\n", m.Status)
			case *message.SensorCommand:
				fmt.Printf("Server command: 0x%02x\n", m.Command)
			default:
				fmt.Printf("Received: %T\n", msg)
			}
		}
	}()

	// Send periodic sensor data
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Send accelerometer data
		sensorData := &message.SensorData{
			SensorID:  1,
			TimeStamp: uint32(time.Now().Unix()),
			Data: message.Data{
				Type:   message.Accelerometer,
				Values: []float32{1.2, -0.5, 9.8},
			},
		}

		if err := conn.Send(sensorData, message.MsgTypeSensorData); err != nil {
			fmt.Printf("Send error: %v\n", err)
			break
		}
		fmt.Printf("Sent sensor data: %v\n", sensorData.Data.Values)

		// Send heartbeat every few cycles
		time.Sleep(1 * time.Second)
		heartbeat := &message.SensorHeartbeat{
			SensorID:  1,
			TimeStamp: uint32(time.Now().Unix()),
			Battery:   85,
			Status:    message.Ok,
		}

		if err := conn.Send(heartbeat, message.MsgTypeHeartbeat); err != nil {
			fmt.Printf("Send error: %v\n", err)
			break
		}
		fmt.Printf("Sent heartbeat: battery=%d%%\n", heartbeat.Battery)
	}
}
