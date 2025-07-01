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
			fmt.Printf("Sensor data from %d: type=%d, values=%v\n", m.SensorID, m.Data.Type, m.Data.Values)

		case *message.SensorDataMulti:
			fmt.Printf("Multi sensor data from %d: %d datasets\n", m.SensorID, len(m.Data))
			for i, data := range m.Data {
				fmt.Printf("  Dataset %d: type=%d, values=%v\n", i+1, data.Type, data.Values)
			}

		case *message.SensorHeartbeat:
			fmt.Printf("Heartbeat from sensor %d, battery: %d%%\n", m.SensorID, m.Battery)

		case *message.RelayedMessage:
			fmt.Printf("Relayed message from relay %d, original data length: %d bytes\n", m.RelayID, len(m.OriginalData))

			originalMsg, err := codec.Unmarshal(m.OriginalData, message.TransportNone)
			if err != nil {
				fmt.Printf("  Failed to decode original message: %v\n", err)
			} else {
				switch orig := originalMsg.(type) {
				case *message.SensorHeartbeat:
					fmt.Printf("  Original: Heartbeat from sensor %d, battery: %d%%\n", orig.SensorID, orig.Battery)
				case *message.SensorData:
					fmt.Printf("  Original: Sensor data from %d, values: %v\n", orig.SensorID, orig.Data.Values)
				default:
					fmt.Printf("  Original: %T\n", originalMsg)
				}
			}

		default:
			fmt.Printf("Received message type: %T\n", msg)
		}
	}
}
