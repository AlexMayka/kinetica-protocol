package main

import (
	"fmt"
	"kinetica-protocol/protocol/codec"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"kinetica-protocol/transport/tcp"
	"log"
	"time"
)

func main() {
	config := &transport.Config{
		Type:           transport.TCP,
		Address:        "localhost:8080",
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    10 * time.Second,
		DefaultCRCType: message.TransportNone,
	}

	tcpTransport := tcp.NewTransportTCP(config)
	defer tcpTransport.Close()

	conn, err := tcpTransport.Connection()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to server")

	registration := &message.Registration{
		SensorID:     1,
		DeviceType:   message.DeviceType9Axis,
		Capabilities: message.CapAccelerometer | message.CapGyroscope,
		FWVersion:    0x0100,
	}

	fmt.Println("Sending registration...")
	if _, err := conn.Send(registration, message.MsgTypeRegister); err != nil {
		log.Fatalf("Failed to send registration: %v", err)
	}

	go func() {
		for {
			msg, err := conn.Receive()
			if err != nil {
				fmt.Printf("Error receiving: %v\n", err)
				return
			}

			switch m := msg.(type) {
			case *message.Ack:
				fmt.Printf("Received ACK: status=%d\n", m.Status)
			case *message.SensorCommand:
				fmt.Printf("Received command: %x\n", m.Command)
			default:
				fmt.Printf("Received: %T\n", msg)
			}
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	messageCount := 0
	for range ticker.C {
		messageCount++

		switch messageCount % 4 {
		case 1:
			sensorData := &message.SensorData{
				SensorID:  1,
				TimeStamp: uint32(time.Now().Unix()),
				Data: message.Data{
					Type:   message.Accelerometer,
					Values: []float32{1.2, -0.5, 9.8},
				},
			}
			fmt.Printf("Sending sensor data: %v\n", sensorData.Data.Values)
			if _, err := conn.Send(sensorData, message.MsgTypeSensorData); err != nil {
				fmt.Printf("Error sending data: %v\n", err)
			}

		case 2:
			sensorDataMulti := &message.SensorDataMulti{
				SensorID:  1,
				TimeStamp: uint32(time.Now().Unix()),
				Data: []message.Data{
					{Type: message.Accelerometer, Values: []float32{1.2, -0.5, 9.8}},
					{Type: message.Gyroscope, Values: []float32{0.1, 0.2, 0.3}},
					{Type: message.Quaternion, Values: []float32{0.0, 0.0, 0.0, 1.0}},
				},
			}
			fmt.Printf("Sending multi sensor data: %d datasets\n", len(sensorDataMulti.Data))
			if _, err := conn.Send(sensorDataMulti, message.MsgTypeSensorDataMulti); err != nil {
				fmt.Printf("Error sending multi data: %v\n", err)
			}

		case 3:
			heartbeat := &message.SensorHeartbeat{
				SensorID:  1,
				TimeStamp: uint32(time.Now().Unix()),
				Battery:   85,
				Status:    message.Ok,
			}
			fmt.Printf("Sending heartbeat: battery=%d%%\n", heartbeat.Battery)
			if _, err := conn.Send(heartbeat, message.MsgTypeHeartbeat); err != nil {
				fmt.Printf("Error sending heartbeat: %v\n", err)
			}

		case 0:
			originalData := &message.SensorHeartbeat{
				SensorID:  5,
				TimeStamp: uint32(time.Now().Unix()),
				Battery:   70,
				Status:    message.Ok,
			}

			originalBytes, _ := codec.Marshal(originalData, 3, message.MsgTypeHeartbeat, message.TransportNone)

			relayedMessage := &message.RelayedMessage{
				RelayID:      10,
				OriginalData: originalBytes,
			}
			fmt.Printf("Sending relayed message from sensor %d via relay %d\n", originalData.SensorID, relayedMessage.RelayID)
			if _, err := conn.Send(relayedMessage, message.MsgTypeRelayed); err != nil {
				fmt.Printf("Error sending relayed message: %v\n", err)
			}
		}
	}
}
