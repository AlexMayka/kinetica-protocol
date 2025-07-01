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
	go runServer()
	time.Sleep(500 * time.Millisecond)
	runClient()
}

func runServer() {
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

	fmt.Printf("🖥️  [%s] Server listening on :8080\n", time.Now().Format("15:04:05.000"))

	for conn := range connections {
		go handleConnection(conn)
	}
}

func handleConnection(conn transport.Connection) {
	defer conn.Close()

	fmt.Printf("🔗 [%s] New connection from %s\n", time.Now().Format("15:04:05.000"), conn.RemoteAddr())

	for {
		msg, err := conn.Receive()
		if err != nil {
			fmt.Printf("❌ [%s] Error receiving message: %v\n", time.Now().Format("15:04:05.000"), err)
			return
		}

		var msgType message.MsgType
		var packetSize int
		switch msg.(type) {
		case *message.Registration:
			msgType = message.MsgTypeRegister
		case *message.SensorData:
			msgType = message.MsgTypeSensorData
		case *message.SensorDataMulti:
			msgType = message.MsgTypeSensorDataMulti
		case *message.SensorHeartbeat:
			msgType = message.MsgTypeHeartbeat
		case *message.RelayedMessage:
			msgType = message.MsgTypeRelayed
		case *message.Ack:
			msgType = message.MsgTypeAck
		default:
			msgType = message.MsgType(0)
		}

		if data, err := codec.Marshal(msg, 1, msgType, message.TransportNone); err == nil {
			packetSize = len(data)
		}

		switch m := msg.(type) {
		case *message.Registration:
			fmt.Printf("📝 [%s] Registration from sensor %d, device type: %d (%d bytes)\n", time.Now().Format("15:04:05.000"), m.SensorID, m.DeviceType, packetSize)
			ack := &message.Ack{
				SensorID:  m.SensorID,
				MessageID: 1,
				Status:    message.AckOK,
			}
			if _, err := conn.Send(ack, message.MsgTypeAck); err != nil {
				fmt.Printf("❌ Error sending ack: %v\n", err)
			} else {
				fmt.Printf("✅ [%s] Sent ACK to sensor %d\n", time.Now().Format("15:04:05.000"), m.SensorID)
			}

		case *message.SensorData:
			fmt.Printf("📊 [%s] Sensor data from %d: type=%d, values=%v (%d bytes)\n", time.Now().Format("15:04:05.000"), m.SensorID, m.Data.Type, m.Data.Values, packetSize)

		case *message.SensorDataMulti:
			fmt.Printf("📊📊 [%s] Multi sensor data from %d: %d datasets (%d bytes)\n", time.Now().Format("15:04:05.000"), m.SensorID, len(m.Data), packetSize)
			for i, data := range m.Data {
				fmt.Printf("     Dataset %d: type=%d, values=%v\n", i+1, data.Type, data.Values)
			}

		case *message.SensorHeartbeat:
			fmt.Printf("💓 [%s] Heartbeat from sensor %d, battery: %d%% (%d bytes)\n", time.Now().Format("15:04:05.000"), m.SensorID, m.Battery, packetSize)

		case *message.RelayedMessage:
			fmt.Printf("🔄 [%s] Relayed message from relay %d, original data: %d bytes (%d bytes total)\n", time.Now().Format("15:04:05.000"), m.RelayID, len(m.OriginalData), packetSize)
			// Декодируем оригинальное сообщение для демонстрации
			if originalMsg, err := codec.Unmarshal(m.OriginalData, message.TransportNone); err == nil {
				switch orig := originalMsg.(type) {
				case *message.SensorHeartbeat:
					fmt.Printf("     Original: Heartbeat from sensor %d, battery: %d%%\n", orig.SensorID, orig.Battery)
				case *message.SensorData:
					fmt.Printf("     Original: Sensor data from %d, values: %v\n", orig.SensorID, orig.Data.Values)
				default:
					fmt.Printf("     Original: %T\n", originalMsg)
				}
			}

		default:
			fmt.Printf("❓ Received message type: %T\n", msg)
		}
	}
}

func runClient() {
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

	fmt.Printf("📱 [%s] Client connected to server\n", time.Now().Format("15:04:05.000"))

	registration := &message.Registration{
		SensorID:     1,
		DeviceType:   message.DeviceType9Axis,
		Capabilities: message.CapAccelerometer | message.CapGyroscope,
		FWVersion:    0x0100,
	}

	fmt.Printf("📤 [%s] Sending registration...\n", time.Now().Format("15:04:05.000"))
	if _, err := conn.Send(registration, message.MsgTypeRegister); err != nil {
		log.Fatalf("Failed to send registration: %v", err)
	}

	go func() {
		for {
			msg, err := conn.Receive()
			if err != nil {
				fmt.Printf("❌ Client error receiving: %v\n", err)
				return
			}

			switch m := msg.(type) {
			case *message.Ack:
				if ackData, err := codec.Marshal(m, 1, message.MsgTypeAck, message.TransportNone); err == nil {
					fmt.Printf("✅ [%s] Client received ACK: status=%d (%d bytes)\n", time.Now().Format("15:04:05.000"), m.Status, len(ackData))
				} else {
					fmt.Printf("✅ [%s] Client received ACK: status=%d\n", time.Now().Format("15:04:05.000"), m.Status)
				}
			default:
				fmt.Printf("📨 Client received: %T\n", msg)
			}
		}
	}()

	for i := 0; i < 6; i++ {
		time.Sleep(1 * time.Second)

		switch i % 4 {
		case 0:
			// Обычные данные сенсора
			sensorData := &message.SensorData{
				SensorID:  1,
				TimeStamp: uint32(time.Now().Unix()),
				Data: message.Data{
					Type:   message.Accelerometer,
					Values: []float32{1.2 + float32(i)*0.1, -0.5, 9.8},
				},
			}
			fmt.Printf("📤 [%s] Sending sensor data #%d: %v\n", time.Now().Format("15:04:05.000"), i+1, sensorData.Data.Values)
			if _, err := conn.Send(sensorData, message.MsgTypeSensorData); err != nil {
				fmt.Printf("❌ Error sending data: %v\n", err)
			}

		case 1:
			// Объединенные данные нескольких сенсоров
			sensorDataMulti := &message.SensorDataMulti{
				SensorID:  1,
				TimeStamp: uint32(time.Now().Unix()),
				Data: []message.Data{
					{Type: message.Accelerometer, Values: []float32{1.2 + float32(i)*0.1, -0.5, 9.8}},
					{Type: message.Gyroscope, Values: []float32{0.1, 0.2 + float32(i)*0.05, 0.3}},
					{Type: message.Quaternion, Values: []float32{0.0, 0.0, 0.0, 1.0}},
				},
			}
			fmt.Printf("📤 [%s] Sending multi sensor data #%d: %d datasets\n", time.Now().Format("15:04:05.000"), i+1, len(sensorDataMulti.Data))
			if _, err := conn.Send(sensorDataMulti, message.MsgTypeSensorDataMulti); err != nil {
				fmt.Printf("❌ Error sending multi data: %v\n", err)
			}

		case 2:
			// Сердцебиение
			heartbeat := &message.SensorHeartbeat{
				SensorID:  1,
				TimeStamp: uint32(time.Now().Unix()),
				Battery:   uint8(85 - i),
				Status:    message.Ok,
			}
			fmt.Printf("📤 [%s] Sending heartbeat #%d: battery=%d%%\n", time.Now().Format("15:04:05.000"), i+1, heartbeat.Battery)
			if _, err := conn.Send(heartbeat, message.MsgTypeHeartbeat); err != nil {
				fmt.Printf("❌ Error sending heartbeat: %v\n", err)
			}

		case 3:
			// Ретранслированное сообщение (имитация ESP-NOW)
			originalData := &message.SensorHeartbeat{
				SensorID:  5,
				TimeStamp: uint32(time.Now().Unix()),
				Battery:   uint8(70 - i),
				Status:    message.Ok,
			}
			originalBytes, _ := codec.Marshal(originalData, 3, message.MsgTypeHeartbeat, message.TransportNone)
			
			relayedMessage := &message.RelayedMessage{
				RelayID:      uint8(10 + i),
				OriginalData: originalBytes,
			}
			fmt.Printf("📤 [%s] Sending relayed message #%d: from sensor %d via relay %d\n", time.Now().Format("15:04:05.000"), i+1, originalData.SensorID, relayedMessage.RelayID)
			if _, err := conn.Send(relayedMessage, message.MsgTypeRelayed); err != nil {
				fmt.Printf("❌ Error sending relayed message: %v\n", err)
			}
		}
	}

	fmt.Printf("🏁 [%s] Client finished sending data\n", time.Now().Format("15:04:05.000"))
	time.Sleep(2 * time.Second)
}
