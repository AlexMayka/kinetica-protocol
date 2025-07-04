// Package main demonstrates a UDP server implementation using the Kinetica protocol.
// This example shows how to receive and process UDP datagrams from sensor clients.
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
	// Configure UDP server
	config := net.Config{
		Address:      ":8082",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  0,
	}

	// Create UDP transport
	transport := net.NewUDP(config)
	defer transport.Close()

	// Start listening for connections
	connections, err := transport.Listen()
	if err != nil {
		log.Fatalf("Failed to start UDP server: %v", err)
	}

	fmt.Println("UDP server listening on :8082")

	// Handle each connection
	for conn := range connections {
		go handleConnection(conn)
	}
}

func handleConnection(conn transport.Connection) {
	defer conn.Close()
	fmt.Println("UDP client connected")

	for {
		// Receive message from client
		msg, err := conn.Receive()
		if err != nil {
			fmt.Printf("UDP client disconnected: %v\n", err)
			return
		}

		// Handle different message types
		switch m := msg.(type) {
		case *message.Registration:
			fmt.Printf("UDP Sensor %d registered (type: %d)\n", m.SensorID, m.DeviceType)

		case *message.SensorData:
			fmt.Printf("UDP Sensor %d data: %v\n", m.SensorID, m.Data.Values)

		case *message.SensorHeartbeat:
			fmt.Printf("UDP Sensor %d heartbeat: %d%% battery\n", m.SensorID, m.Battery)

		default:
			fmt.Printf("UDP unknown message: %T\n", msg)
		}
	}
}