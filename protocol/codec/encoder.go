package codec

import (
	"encoding/binary"
	"fmt"
	"kinetica-protocol/protocol/message"
)

func (buf *buffer) encodePayload(msg message.Message) error {
	switch m := msg.(type) {
	case *message.SensorCommand:
		return buf.encodeCommand(m)
	case *message.SensorConfig:
		return buf.encodeConfig(m)
	case *message.SensorHeartbeat:
		return buf.encodeHeartbeat(m)
	case *message.SensorData:
		return buf.encodeData(m)
	case *message.CustomData:
		return buf.encodeCustomData(m)
	case *message.TimeSync:
		return buf.encodeTimeSync(m)
	case *message.Ack:
		return buf.encodeAck(m)
	case *message.Registration:
		return buf.encodeRegistration(m)
	case *message.Fragment:
		return buf.encodeFragment(m)
	case *message.RelayedMessage:
		return buf.encodeRelayedMessage(m)
	case *message.SensorDataMulti:
		return buf.encodeDataMulti(m)
	default:
		return ErrInvalidMessageType
	}
}

func (buf *buffer) encodeHeader(packetID uint8, msgType message.MsgType, payloadLength uint8) error {
	header := message.NewHeader(packetID, msgType, payloadLength)
	if err := binary.Write(buf.bufHeader, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("%w: header encoding failed", ErrEncodingFailed)
	}

	return nil
}

func (buf *buffer) encodeFooter(transportType message.TransportCRC) error {
	data := buf.bytes()
	footer := message.NewFooter(transportType, data)

	if err := binary.Write(buf.bufFooter, binary.LittleEndian, footer.Bytes); err != nil {
		return fmt.Errorf("%w: footer encoding failed", ErrEncodingFailed)
	}

	return nil
}

func (buf *buffer) writeField(field interface{}, fieldName string) error {
	if err := binary.Write(buf.bufPayload, binary.LittleEndian, field); err != nil {
		return fmt.Errorf("%w: %s encoding failed", ErrEncodingFailed, fieldName)
	}
	return nil
}

func (buf *buffer) appendItem(item message.Item) error {
	buf.bufPayload.WriteByte(uint8(item.Key))
	buf.bufPayload.WriteByte(item.Length)

	if _, err := buf.bufPayload.Write(item.Value); err != nil {
		return fmt.Errorf("%w: item value encoding failed", ErrEncodingFailed)
	}

	return nil
}

func (buf *buffer) encodeCommand(msg *message.SensorCommand) error {
	buf.bufPayload.WriteByte(msg.SensorID)

	if err := buf.writeField(msg.TimeStamp, "command timestamp"); err != nil {
		return err
	}

	buf.bufPayload.WriteByte(msg.Command)
	return nil
}

func (buf *buffer) encodeConfig(msg *message.SensorConfig) error {
	buf.bufPayload.WriteByte(msg.SensorID)
	if err := buf.writeField(msg.TimeStamp, "config timestamp"); err != nil {
		return err
	}

	buf.bufPayload.WriteByte(uint8(len(msg.Config)))

	for _, item := range msg.Config {
		if err := buf.appendItem(item); err != nil {
			return fmt.Errorf("%w: config item encoding failed", ErrEncodingFailed)
		}
	}

	return nil
}

func (buf *buffer) encodeHeartbeat(msg *message.SensorHeartbeat) error {
	buf.bufPayload.WriteByte(msg.SensorID)

	if err := buf.writeField(msg.TimeStamp, "heartbeat timestamp"); err != nil {
		return err
	}

	buf.bufPayload.WriteByte(msg.Battery)
	buf.bufPayload.WriteByte(uint8(msg.Status))
	return nil
}

func (buf *buffer) encodeData(msg *message.SensorData) error {
	buf.bufPayload.WriteByte(msg.SensorID)
	if err := buf.writeField(msg.TimeStamp, "sensor data timestamp"); err != nil {
		return err
	}

	buf.bufPayload.WriteByte(uint8(msg.Data.Type))
	buf.bufPayload.WriteByte(uint8(len(msg.Data.Values)))

	for _, value := range msg.Data.Values {
		if err := buf.writeField(value, "sensor value"); err != nil {
			return err
		}
	}

	return nil
}

func (buf *buffer) encodeCustomData(msg *message.CustomData) error {
	buf.bufPayload.WriteByte(msg.SensorID)
	if err := buf.writeField(msg.TimeStamp, "custom data timestamp"); err != nil {
		return err
	}

	buf.bufPayload.WriteByte(uint8(msg.DataType))
	buf.bufPayload.WriteByte(uint8(len(msg.Data)))

	for _, item := range msg.Data {
		err := buf.appendItem(item)
		if err != nil {
			return fmt.Errorf("%w: custom data item encoding failed", ErrEncodingFailed)
		}
	}

	return nil
}

func (buf *buffer) encodeTimeSync(msg *message.TimeSync) error {
	buf.bufPayload.WriteByte(msg.SensorID)

	if err := buf.writeField(msg.ServerTime, "server time"); err != nil {
		return err
	}
	if err := buf.writeField(msg.SensorTime, "sensor time"); err != nil {
		return err
	}

	return nil
}

func (buf *buffer) encodeAck(msg *message.Ack) error {
	buf.bufPayload.WriteByte(msg.SensorID)

	if err := buf.writeField(msg.MessageID, "ack message ID"); err != nil {
		return err
	}

	buf.bufPayload.WriteByte(uint8(msg.Status))
	return nil
}

func (buf *buffer) encodeRegistration(msg *message.Registration) error {
	buf.bufPayload.WriteByte(msg.SensorID)
	buf.bufPayload.WriteByte(uint8(msg.DeviceType))
	buf.bufPayload.WriteByte(msg.Capabilities)

	if err := buf.writeField(msg.FWVersion, "firmware version"); err != nil {
		return err
	}

	return nil
}

func (buf *buffer) encodeFragment(msg *message.Fragment) error {
	if err := buf.writeField(msg.MessageID, "fragment message ID"); err != nil {
		return err
	}

	buf.bufPayload.WriteByte(msg.FragmentNum)
	buf.bufPayload.WriteByte(msg.TotalFragments)

	if err := buf.writeField(uint16(len(msg.Data)), "fragment data length"); err != nil {
		return err
	}
	if _, err := buf.bufPayload.Write(msg.Data); err != nil {
		return fmt.Errorf("%w: fragment data encoding failed", ErrEncodingFailed)
	}

	return nil
}

func (buf *buffer) encodeRelayedMessage(msg *message.RelayedMessage) error {
	buf.bufPayload.WriteByte(msg.RelayID)

	if err := buf.writeField(uint16(len(msg.OriginalData)), "relayed data length"); err != nil {
		return err
	}

	if err := buf.writeField(msg.OriginalData, "relayed data encoding failed"); err != nil {
		return err
	}

	return nil
}

func (buf *buffer) encodeDataMulti(m *message.SensorDataMulti) error {
	buf.bufPayload.WriteByte(m.SensorID)

	if err := buf.writeField(m.TimeStamp, "timestamp"); err != nil {
		return err
	}

	buf.bufPayload.WriteByte(uint8(len(m.Data)))

	for _, data := range m.Data {
		if err := buf.writeField(data.Type, "data"); err != nil {
			return err
		}

		buf.bufPayload.WriteByte(uint8(len(data.Values)))
		for _, value := range data.Values {
			if err := buf.writeField(value, "data"); err != nil {
				return err
			}
		}
	}

	return nil
}
