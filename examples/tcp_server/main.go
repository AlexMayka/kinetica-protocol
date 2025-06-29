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
		Address:        ":8080",
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    0,
		DefaultCRCType: message.TransportNone,
	}

	tcpTransport := tcp.NewTransportTCP(config)
	defer tcpTransport.Close()

	connections, err := tcpTransport.Listen()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	fmt.Println("Server listening on :8080")

	for conn := range connections {
		go handleConnection(conn)
	}
}

func handleConnection(conn transport.Connection) {
	defer conn.Close()

	fmt.Printf("New connection from %s\n", conn.RemoteAddr())

	for {
		msg, err := conn.Receive()
		if err != nil {
			fmt.Printf("Error receiving message: %v\n", err)
			return
		}

		switch m := msg.(type) {
		case *message.Registration:
			fmt.Printf("Registration from sensor %d, device type: %d\n", m.SensorID, m.DeviceType)
			ack := &message.Ack{
				SensorID:  m.SensorID,
				MessageID: 1,
				Status:    message.AckOK,
			}
			if _, err := conn.Send(ack, message.MsgTypeAck); err != nil {
				fmt.Printf("Error sending ack: %v\n", err)
			}

		case *message.SensorData:
			fmt.Printf("Sensor data from %d: %v\n", m.SensorID, m.Values)

		case *message.SensorHeartbeat:
			fmt.Printf("Heartbeat from sensor %d, battery: %d%%\n", m.SensorID, m.Battery)

		default:
			fmt.Printf("Received message type: %T\n", msg)
		}
	}
}
