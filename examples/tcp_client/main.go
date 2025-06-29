package main

import (
	"fmt"
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

	for range ticker.C {
		sensorData := &message.SensorData{
			SensorID:  1,
			TimeStamp: uint32(time.Now().Unix()),
			Type:      message.Accelerometer,
			Values:    []float32{1.2, -0.5, 9.8},
		}

		fmt.Printf("Sending sensor data: %v\n", sensorData.Values)
		if _, err := conn.Send(sensorData, message.MsgTypeSensorData); err != nil {
			fmt.Printf("Error sending data: %v\n", err)
		}

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
	}
}
