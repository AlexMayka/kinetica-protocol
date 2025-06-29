package codec

import (
	"bytes"
	"fmt"
	"kinetica-protocol/protocol/message"
)

type buffer struct {
	bufHeader  *bytes.Buffer
	bufPayload *bytes.Buffer
	bufFooter  *bytes.Buffer
}

func newBuffer() *buffer {
	return &buffer{
		bufHeader:  new(bytes.Buffer),
		bufPayload: new(bytes.Buffer),
		bufFooter:  new(bytes.Buffer),
	}
}

func (buf *buffer) bytes() []byte {
	result := make([]byte, 0, buf.bufHeader.Len()+buf.bufPayload.Len()+buf.bufFooter.Len())
	result = append(result, buf.bufHeader.Bytes()...)
	result = append(result, buf.bufPayload.Bytes()...)
	result = append(result, buf.bufFooter.Bytes()...)
	return result
}

type packet struct {
	buf          *bytes.Buffer
	originalData []byte
	header       message.Header
	payload      message.Message
	footer       message.Footer
}

func newPacket(data []byte) *packet {
	return &packet{
		buf:          bytes.NewBuffer(data),
		originalData: data,
		header:       message.Header{},
		footer:       message.Footer{},
	}
}

func Marshal(msg message.Message, msgType message.MsgType, transportType message.TransportCRC) ([]byte, error) {
	buf := newBuffer()

	err := buf.encodePayload(msg)
	if err != nil {
		return nil, err
	}

	err = buf.encodeHeader(0, msgType, uint8(len(buf.bufPayload.Bytes())))
	if err != nil {
		return nil, err
	}

	err = buf.encodeFooter(transportType)
	if err != nil {
		return nil, err
	}

	return buf.bytes(), nil
}

func Unmarshal(data []byte, transport message.TransportCRC) (message.Message, error) {
	if len(data) < message.HeaderSize {
		return nil, fmt.Errorf("%w: need at least %d bytes", ErrMessageTooShort, message.HeaderSize)
	}

	p := newPacket(data)
	if err := p.decodeHeader(); err != nil {
		return nil, err
	}

	if p.header.Magic != message.MagicBytes {
		return nil, fmt.Errorf("%w: got %x", ErrInvalidMagicBytes, p.header.Magic)
	}

	if err := p.decodePayload(); err != nil {
		return nil, err
	}

	if err := p.validateFooter(transport); err != nil {
		return nil, err
	}

	return p.payload, nil
}
