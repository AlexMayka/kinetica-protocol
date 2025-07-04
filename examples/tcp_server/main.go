// Package main demonstrates a TCP server implementation using the Kinetica protocol.
// This example shows how to start a TCP server, accept client connections,
// and handle different types of sensor messages.
package main

import (
	"fmt"
	"kinetica-protocol/protocol/message"
	"kinetica-protocol/transport"
	"kinetica-protocol/transport/net"
	"log"
	"time"
)

func main() {
	// Configure TCP server
	config := net.Config{
		Address:      ":8081",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  0, // No read timeout for server
	}

	// Create TCP transport
	transport := net.NewTCP(config)
	defer transport.Close()

	// Start listening for connections
	connections, err := transport.Listen()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Println("Server listening on :8081")

	// Handle each connection in a separate goroutine
	for conn := range connections {
		go handleConnection(conn)
	}
}

// handleConnection processes messages from a single client connection.
// It handles sensor registration, data messages, and heartbeats,
// sending appropriate acknowledgments back to the client.
func handleConnection(conn transport.Connection) {
	defer conn.Close()
	fmt.Println("Client connected")

	for {
		// Receive message from client
		msg, err := conn.Receive()
		if err != nil {
			fmt.Printf("Client disconnected: %v\n", err)
			return
		}

		// Handle different message types
		switch m := msg.(type) {
		case *message.Registration:
			fmt.Printf("Sensor %d registered (type: %d, capabilities: 0x%02x)\n", 
				m.SensorID, m.DeviceType, m.Capabilities)
			
			// Send acknowledgment
			ack := &message.Ack{
				SensorID:  m.SensorID,
				MessageID: 1,
				Status:    message.AckOK,
			}
			if err := conn.Send(ack, message.MsgTypeAck); err != nil {
				fmt.Printf("Failed to send ACK: %v\n", err)
			}

		case *message.SensorData:
			fmt.Printf("Sensor %d data: %v\n", m.SensorID, m.Data.Values)

		case *message.SensorHeartbeat:
			fmt.Printf("Sensor %d heartbeat: %d%% battery\n", m.SensorID, m.Battery)

		default:
			fmt.Printf("Unknown message type: %T\n", msg)
		}
	}
}
