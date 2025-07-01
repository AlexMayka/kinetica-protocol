package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"kinetica-protocol/protocol/message"
)

func (p *packet) decodeHeader() error {
	if err := binary.Read(p.buf, binary.LittleEndian, &p.header); err != nil {
		return err
	}
	return nil
}

func (p *packet) decodePayload() error {
	payloadBytes := make([]byte, p.header.Length)

	n, err := p.buf.Read(payloadBytes)
	if err != nil || n != int(p.header.Length) {
		return err
	}

	buf := bytes.NewBuffer(payloadBytes)

	switch p.header.Type {
	case message.MsgTypeCommand:
		return p.decodeCommand(buf)
	case message.MsgTypeConfig:
		return p.decodeConfig(buf)
	case message.MsgTypeHeartbeat:
		return p.decodeHeartbeat(buf)
	case message.MsgTypeSensorData:
		return p.decodeData(buf)
	case message.MsgTypeCustom:
		return p.decodeCustomData(buf)
	case message.MsgTypeTimeSync:
		return p.decodeTimeSync(buf)
	case message.MsgTypeAck:
		return p.decodeAck(buf)
	case message.MsgTypeRegister:
		return p.decodeRegistration(buf)
	case message.MsgTypeFragment:
		return p.decodeFragment(buf)
	case message.MsgTypeRelayed:
		return p.decodeRelayedMessage(buf)
	case message.MsgTypeSensorDataMulti:
		return p.decodeDataMulti(buf)
	}

	return nil
}

func (p *packet) readField(buf *bytes.Buffer, field interface{}, fieldName string) error {
	if err := binary.Read(buf, binary.LittleEndian, field); err != nil {
		return fmt.Errorf("%w: failed to read %s", ErrDecodingFailed, fieldName)
	}
	return nil
}

func (p *packet) decodeItem(buf *bytes.Buffer) (*message.Item, error) {
	item := &message.Item{}

	if err := p.readField(buf, &item.Key, "item key"); err != nil {
		return nil, err
	}
	if err := p.readField(buf, &item.Length, "item length"); err != nil {
		return nil, err
	}

	item.Value = make([]byte, item.Length)
	if _, err := buf.Read(item.Value); err != nil {
		return nil, fmt.Errorf("%w: failed to read item value", ErrDecodingFailed)
	}

	return item, nil
}

func (p *packet) validateFooter(transport message.TransportCRC) error {
	if transport == message.TransportNone {
		return nil
	}

	footerSize := message.GetFooterSize(transport)
	if p.buf.Len() < footerSize {
		return fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidFooter, footerSize, p.buf.Len())
	}

	receivedFooter := make([]byte, footerSize)
	if _, err := p.buf.Read(receivedFooter); err != nil {
		return fmt.Errorf("%w: failed to read footer", ErrDecodingFailed)
	}

	totalDataSize := message.HeaderSize + int(p.header.Length)
	if len(p.originalData) < totalDataSize {
		return fmt.Errorf("%w: original data too short for validation", ErrInsufficientData)
	}

	dataForValidation := p.originalData[:totalDataSize]

	expectedFooter := message.NewFooter(transport, dataForValidation)

	if !bytes.Equal(receivedFooter, expectedFooter.Bytes) {
		return fmt.Errorf("%w: expected %x, got %x", ErrInvalidFooter, expectedFooter.Bytes, receivedFooter)
	}

	return nil
}

func (p *packet) decodeCommand(buf *bytes.Buffer) error {
	data := message.SensorCommand{}

	if err := p.readField(buf, &data.SensorID, "sensor ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.TimeStamp, "timestamp"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.Command, "command"); err != nil {
		return err
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeConfig(buf *bytes.Buffer) error {
	data := message.SensorConfig{}

	if err := p.readField(buf, &data.SensorID, "sensor ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.TimeStamp, "timestamp"); err != nil {
		return err
	}

	var countItems uint8
	if err := p.readField(buf, &countItems, "config items count"); err != nil {
		return err
	}

	for i := 0; i < int(countItems); i++ {
		value, err := p.decodeItem(buf)
		if err != nil {
			return err
		}
		data.Config = append(data.Config, *value)
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeHeartbeat(buf *bytes.Buffer) error {
	data := message.SensorHeartbeat{}

	if err := p.readField(buf, &data.SensorID, "sensor ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.TimeStamp, "timestamp"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.Battery, "battery level"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.Status, "status"); err != nil {
		return err
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeData(buf *bytes.Buffer) error {
	data := message.SensorData{}

	if err := p.readField(buf, &data.SensorID, "sensor ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.TimeStamp, "timestamp"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.Data.Type, "data type"); err != nil {
		return err
	}

	var countItem uint8
	if err := p.readField(buf, &countItem, "values count"); err != nil {
		return err
	}

	for i := 0; i < int(countItem); i++ {
		var item float32
		if err := p.readField(buf, &item, "sensor value"); err != nil {
			return err
		}
		data.Data.Values = append(data.Data.Values, item)
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeCustomData(buf *bytes.Buffer) error {
	data := message.CustomData{}

	if err := p.readField(buf, &data.SensorID, "sensor ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.TimeStamp, "timestamp"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.DataType, "data type"); err != nil {
		return err
	}

	var countItems uint8
	if err := p.readField(buf, &countItems, "custom items count"); err != nil {
		return err
	}

	for i := 0; i < int(countItems); i++ {
		value, err := p.decodeItem(buf)
		if err != nil {
			return err
		}
		data.Data = append(data.Data, *value)
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeTimeSync(buf *bytes.Buffer) error {
	data := message.TimeSync{}

	if err := p.readField(buf, &data.SensorID, "sensor ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.ServerTime, "server time"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.SensorTime, "sensor time"); err != nil {
		return err
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeAck(buf *bytes.Buffer) error {
	data := message.Ack{}

	if err := p.readField(buf, &data.SensorID, "sensor ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.MessageID, "message ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.Status, "status"); err != nil {
		return err
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeRegistration(buf *bytes.Buffer) error {
	data := message.Registration{}

	if err := p.readField(buf, &data.SensorID, "sensor ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.DeviceType, "device type"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.Capabilities, "capabilities"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.FWVersion, "firmware version"); err != nil {
		return err
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeFragment(buf *bytes.Buffer) error {
	data := message.Fragment{}

	if err := p.readField(buf, &data.MessageID, "message ID"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.FragmentNum, "fragment number"); err != nil {
		return err
	}
	if err := p.readField(buf, &data.TotalFragments, "total fragments"); err != nil {
		return err
	}

	var dataLength uint16
	if err := p.readField(buf, &dataLength, "data length"); err != nil {
		return err
	}

	data.Data = make([]byte, dataLength)
	if _, err := buf.Read(data.Data); err != nil {
		return err
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeRelayedMessage(buf *bytes.Buffer) error {
	data := message.RelayedMessage{}

	if err := p.readField(buf, &data.RelayID, "relay ID"); err != nil {
		return err
	}

	var dataLength uint16
	if err := p.readField(buf, &dataLength, "original data length"); err != nil {
		return err
	}

	data.OriginalData = make([]byte, dataLength)
	if _, err := buf.Read(data.OriginalData); err != nil {
		return fmt.Errorf("%w: failed to read original data", ErrDecodingFailed)
	}

	p.payload = &data
	return nil
}

func (p *packet) decodeDataMulti(buf *bytes.Buffer) error {
	data := message.SensorDataMulti{}

	if err := p.readField(buf, &data.SensorID, "sensor ID"); err != nil {
		return err
	}

	if err := p.readField(buf, &data.TimeStamp, "timestamp"); err != nil {
		return err
	}

	var lengthData uint8
	if err := p.readField(buf, &lengthData, "length"); err != nil {
		return err
	}

	data.Data = make([]message.Data, lengthData)

	for i := 0; i < int(lengthData); i++ {
		if err := p.readField(buf, &data.Data[i].Type, "type data"); err != nil {
			return err
		}

		var lengthItem uint8
		if err := p.readField(buf, &lengthItem, fmt.Sprintf("length elem %v", i)); err != nil {
			return err
		}

		for j := 0; j < int(lengthItem); j++ {
			var item float32
			if err := p.readField(buf, &item, "sensor value"); err != nil {
				return err
			}
			data.Data[i].Values = append(data.Data[i].Values, item)
		}
	}

	p.payload = &data
	return nil
}
