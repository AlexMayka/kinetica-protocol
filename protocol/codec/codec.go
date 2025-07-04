// Package codec provides encoding and decoding functionality for the Kinetica protocol.
// It handles the binary serialization of protocol messages including header, payload,
// and footer with CRC validation. The codec supports all message types defined in
// the protocol specification and multiple transport CRC options.
package codec

import (
	"bytes"
	"fmt"
	"kinetica-protocol/protocol/message"
)

// buffer manages separate buffers for header, payload, and footer during encoding.
type buffer struct {
	bufHeader  *bytes.Buffer // Protocol header buffer
	bufPayload *bytes.Buffer // Message payload buffer  
	bufFooter  *bytes.Buffer // CRC footer buffer
}

// newBuffer creates a new buffer with initialized sub-buffers for encoding.
func newBuffer() *buffer {
	return &buffer{
		bufHeader:  new(bytes.Buffer),
		bufPayload: new(bytes.Buffer),
		bufFooter:  new(bytes.Buffer),
	}
}

// bytes concatenates all buffer segments into a complete binary packet.
func (buf *buffer) bytes() []byte {
	result := make([]byte, 0, buf.bufHeader.Len()+buf.bufPayload.Len()+buf.bufFooter.Len())
	result = append(result, buf.bufHeader.Bytes()...)
	result = append(result, buf.bufPayload.Bytes()...)
	result = append(result, buf.bufFooter.Bytes()...)
	return result
}

// packet manages the decoding state for parsing incoming binary data.
type packet struct {
	buf          *bytes.Buffer    // Reading buffer for binary data
	originalData []byte           // Original packet data for CRC validation
	header       message.Header   // Decoded protocol header
	payload      message.Message  // Decoded message payload
	footer       message.Footer   // Decoded footer (if present)
}

// newPacket creates a new packet decoder for the given binary data.
func newPacket(data []byte) *packet {
	return &packet{
		buf:          bytes.NewBuffer(data),
		originalData: data,
		header:       message.Header{},
		footer:       message.Footer{},
	}
}

// Marshal encodes a protocol message into binary format with the specified parameters.
// It creates a complete packet with header, payload, and footer including CRC validation.
//
// Parameters:
//   - msg: The message to encode (must implement message.Message interface)
//   - packetID: Unique packet identifier (0-255, wraps around)
//   - msgType: The type of message being encoded
//   - transportType: The CRC type to use for footer validation
//
// Returns the complete binary packet or an error if encoding fails.
func Marshal(msg message.Message, packetID uint8, msgType message.MsgType, transportType message.TransportCRC) ([]byte, error) {
	buf := newBuffer()

	err := buf.encodePayload(msg)
	if err != nil {
		return nil, err
	}

	err = buf.encodeHeader(packetID, msgType, uint8(len(buf.bufPayload.Bytes())))
	if err != nil {
		return nil, err
	}

	err = buf.encodeFooter(transportType)
	if err != nil {
		return nil, err
	}

	return buf.bytes(), nil
}

// Unmarshal decodes binary data into a protocol message with CRC validation.
// It parses the packet header, validates the magic bytes, decodes the payload,
// and verifies the footer CRC according to the specified transport type.
//
// Parameters:
//   - data: The binary packet data to decode
//   - transport: The expected CRC type for footer validation
//
// Returns the decoded message or an error if validation or decoding fails.
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
