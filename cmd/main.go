package main

import (
	"fmt"
	"kinetica-protocol/protocol/codec"
	"kinetica-protocol/protocol/message"
)

func main() {
	// 1. SensorCommand
	command := &message.SensorCommand{
		SensorID:  1,
		TimeStamp: 12345,
		Command:   0x10,
	}
	testMessage(command, message.MsgTypeCommand, "SensorCommand")

	// 2. SensorConfig
	config := &message.SensorConfig{
		SensorID:  1,
		TimeStamp: 12345,
		Config: []message.Item{
			{Key: message.ConfigKeySampleRate, Length: 4, Value: []byte{0x64, 0x00, 0x00, 0x00}},
			{Key: message.ConfigKeyRange, Length: 2, Value: []byte{0x08, 0x00}},
		},
	}
	testMessage(config, message.MsgTypeConfig, "SensorConfig")

	// 3. SensorHeartbeat
	heartbeat := &message.SensorHeartbeat{
		SensorID:  1,
		TimeStamp: 12345,
		Battery:   85,
		Status:    message.Ok,
	}
	testMessage(heartbeat, message.MsgTypeHeartbeat, "SensorHeartbeat")

	// 4. SensorData
	sensorData := &message.SensorData{
		SensorID:  1,
		TimeStamp: 12345,
		Type:      message.Accelerometer,
		Values:    []float32{1.2, 3.4, 5.6},
	}
	testMessage(sensorData, message.MsgTypeSensorData, "SensorData")

	// 5. CustomData
	customData := &message.CustomData{
		SensorID:  1,
		TimeStamp: 12345,
		DataType:  message.CustomTypeLog,
		Data: []message.Item{
			{Key: message.ConfigKeyDeviceName, Length: 5, Value: []byte("Test1")},
		},
	}
	testMessage(customData, message.MsgTypeCustom, "CustomData")

	// 6. TimeSync
	timeSync := &message.TimeSync{
		SensorID:   1,
		ServerTime: 1234567890,
		SensorTime: 1234567880,
	}
	testMessage(timeSync, message.MsgTypeTimeSync, "TimeSync")

	// 7. Ack
	ack := &message.Ack{
		SensorID:  1,
		MessageID: 123,
		Status:    message.AckOK,
	}
	testMessage(ack, message.MsgTypeAck, "Ack")

	// 8. Registration
	registration := &message.Registration{
		SensorID:     1,
		DeviceType:   message.DeviceTypeBNO085,
		Capabilities: message.CapAccelerometer | message.CapGyroscope,
		FWVersion:    0x0102,
	}
	testMessage(registration, message.MsgTypeRegister, "Registration")

	// 9. Fragment
	fragment := &message.Fragment{
		MessageID:      456,
		FragmentNum:    1,
		TotalFragments: 3,
		Data:           []byte{0x01, 0x02, 0x03, 0x04, 0x05},
	}
	testMessage(fragment, message.MsgTypeFragment, "Fragment")

	// Тест декодирования известного пакета
	fmt.Printf("\n=== Тест декодирования ===\n")
	data := []byte{0x4b, 0x4e, 0x00, 0x01, 0x04, 0x13, 0x01, 0x39, 0x30, 0x00, 0x00, 0x01, 0x03, 0x9a, 0x99, 0x99, 0x3f, 0x9a, 0x99, 0x59, 0x40, 0x33, 0x33, 0xb3, 0x40, 0x30}
	decoded, err := codec.Unmarshal(data, message.TransportCRC8)
	if err != nil {
		fmt.Printf("Ошибка декодирования: %v\n", err)
	} else {
		fmt.Printf("Декодировано: %+v\n", decoded)
	}
}

func testMessage(msg message.Message, msgType message.MsgType, name string) {
	fmt.Printf("\n=== %s ===\n", name)

	fmt.Printf("Изначальное: %+v\n", msg)

	// Кодируем
	data, err := codec.Marshal(msg, msgType, message.TransportCRC8)
	if err != nil {
		fmt.Printf("Ошибка кодирования: %v\n", err)
		return
	}
	fmt.Printf("Закодировано: %x\n", data)

	decoded, err := codec.Unmarshal(data, message.TransportCRC8)
	if err != nil {
		fmt.Printf("Ошибка декодирования: %v\n", err)
		return
	}

	fmt.Printf("Декодировано: %+v\n", decoded)
}
