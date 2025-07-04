// Package main demonstrates a BLE (Bluetooth Low Energy) client implementation using the Kinetica protocol.
// This example shows how to connect to ESP32 devices via BLE GATT services and characteristics.
package main

import (
	"fmt"
	"kinetica-protocol/protocol/message"
	bleTransport "kinetica-protocol/transport/ble"
	"log"
	"time"
	"tinygo.org/x/bluetooth"
)

func main() {
	// Configure BLE connection
	deviceAddr := "AA:BB:CC:DD:EE:FF" // Change to your ESP32 MAC address
	config := bleTransport.Config{
		DeviceAddress:  &deviceAddr,
		ServiceUUID:    bluetooth.New16BitUUID(0x180F), // Battery Service
		WriteCharUUID:  bluetooth.New16BitUUID(0x2A19), // Battery Level
		NotifyCharUUID: bluetooth.New16BitUUID(0x2A1A), // Custom characteristic
		ScanTimeout:    10 * time.Second,
		ReadTimeout:    5 * time.Second,
	}

	// Create BLE transport
	transport := bleTransport.NewBLE(config)
	defer transport.Close()

	fmt.Printf("Scanning for BLE device: %s\n", deviceAddr)

	// Connect to BLE device
	conn, err := transport.Connection()
	if err != nil {
		log.Fatalf("Failed to connect to BLE device: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to BLE device")

	// Register sensor
	registration := &message.Registration{
		SensorID:     4,
		DeviceType:   message.DeviceType9Axis,
		Capabilities: message.CapAccelerometer | message.CapGyroscope,
		FWVersion:    0x0103,
	}

	if err := conn.Send(registration, message.MsgTypeRegister); err != nil {
		log.Fatalf("Failed to send BLE registration: %v", err)
	}
	fmt.Println("Registration sent via BLE")

	// Listen for BLE notifications
	go func() {
		for {
			msg, err := conn.Receive()
			if err != nil {
				fmt.Printf("BLE receive error: %v\n", err)
				return
			}

			switch m := msg.(type) {
			case *message.Ack:
				fmt.Printf("BLE ACK received: status=%d\n", m.Status)
			case *message.SensorCommand:
				fmt.Printf("BLE command received: 0x%02x\n", m.Command)
			default:
				fmt.Printf("BLE received: %T\n", msg)
			}
		}
	}()

	// Send periodic data (small packets for BLE)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Send compact sensor data
		sensorData := &message.SensorData{
			SensorID:  4,
			TimeStamp: uint32(time.Now().Unix()),
			Data: message.Data{
				Type:   message.Accelerometer,
				Values: []float32{0.1, 0.2, 9.8},
			},
		}

		if err := conn.Send(sensorData, message.MsgTypeSensorData); err != nil {
			fmt.Printf("BLE send error: %v\n", err)
			break
		}
		fmt.Printf("Sent BLE data: %v\n", sensorData.Data.Values)

		// Send heartbeat
		time.Sleep(2 * time.Second)
		heartbeat := &message.SensorHeartbeat{
			SensorID:  4,
			TimeStamp: uint32(time.Now().Unix()),
			Battery:   78,
			Status:    message.Ok,
		}

		if err := conn.Send(heartbeat, message.MsgTypeHeartbeat); err != nil {
			fmt.Printf("BLE send error: %v\n", err)
			break
		}
		fmt.Printf("Sent BLE heartbeat: %d%%\n", heartbeat.Battery)
	}
}