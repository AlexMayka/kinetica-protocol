// Package main demonstrates a serial/UART client implementation using the Kinetica protocol.
// This example shows communication over RS232/UART with embedded devices like ESP32.
package main

import (
	"fmt"
	"go.bug.st/serial"
	"kinetica-protocol/protocol/message"
	serialTransport "kinetica-protocol/transport/serial"
	"log"
	"time"
)

func main() {
	// Configure serial connection
	config := serialTransport.Config{
		Port:     "/dev/ttyUSB0", // Change to your serial port
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
		ReadTimeout: 5 * time.Second,
	}

	// Create serial transport
	transport := serialTransport.NewSerial(config)
	defer transport.Close()

	// Connect to serial device
	conn, err := transport.Connection()
	if err != nil {
		log.Fatalf("Failed to open serial port: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to serial port: %s\n", config.Port)

	// Register sensor
	registration := &message.Registration{
		SensorID:     3,
		DeviceType:   message.DeviceType3Axis,
		Capabilities: message.CapAccelerometer,
		FWVersion:    0x0102,
	}

	if err := conn.Send(registration, message.MsgTypeRegister); err != nil {
		log.Fatalf("Failed to send registration: %v", err)
	}
	fmt.Println("Registration sent via serial")

	// Send periodic data
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Send sensor data
		sensorData := &message.SensorData{
			SensorID:  3,
			TimeStamp: uint32(time.Now().Unix()),
			Data: message.Data{
				Type:   message.Accelerometer,
				Values: []float32{2.1, -1.5, 8.9},
			},
		}

		if err := conn.Send(sensorData, message.MsgTypeSensorData); err != nil {
			fmt.Printf("Serial send error: %v\n", err)
			break
		}
		fmt.Printf("Sent serial data: %v\n", sensorData.Data.Values)

		// Send heartbeat
		time.Sleep(500 * time.Millisecond)
		heartbeat := &message.SensorHeartbeat{
			SensorID:  3,
			TimeStamp: uint32(time.Now().Unix()),
			Battery:   92,
			Status:    message.Ok,
		}

		if err := conn.Send(heartbeat, message.MsgTypeHeartbeat); err != nil {
			fmt.Printf("Serial send error: %v\n", err)
			break
		}
		fmt.Printf("Sent serial heartbeat: %d%%\n", heartbeat.Battery)
	}
}