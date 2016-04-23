package websocket

import (
	"bufio"
	"encoding/binary"
	"fmt"
)

/*
	WebSocket Opcodes.
	Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.2
*/
const (
	OpcodeContinuation int = 0
	OpcodeText         int = 1
	OpcodeBinary       int = 2
	OpcodeClose        int = 8
	OpcodePing         int = 9
	OpcodePong         int = 10
)

/*
	frame represents a Websocket Data Frame.
	Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.2
*/
type frame struct {
	/*
		fin indicates that the frame is the final fragment.
	*/
	fin bool

	/*
		opcode defines the interpretation of the payload data.
	*/
	opcode int

	/*
		masked defines whether the payload data is masked.
	*/
	masked bool

	/*
		length specifies the length of the payload data in bytes.
	*/
	length uint64

	/*
		key contains the masking key to be used to decode the payload data (if
		data is masked). It is 32 bits in length.
	*/
	key []byte

	/*
		payload contains the data received from the client.
	*/
	payload []byte
}

/*
	newFrame is a constructor function to create a new instance of frame by
	reading from a buffer. The construction of the websocket frame is divided
	into four sections:
		1. Parsing of first 2 bytes.
		2. Parsing of 'payload length' if 'payload length' parsed in first
		   section is greater 125.
		3. Parsing of 'masking key' if 'masked' value parsed in first section is
		   set to true.
		4. Parsing of payload data.
*/
func newFrame(b *bufio.Reader) (*frame, error) {
	// Create frame instance.
	f := &frame{}

	reads := []func(*bufio.Reader) error{
		f.readInitial,
		f.readLength,
		f.readMaskKey,
		f.readPayload,
	}

	for _, read := range reads {
		if err := read(b); err != nil {
			return nil, err
		}
	}

	return f, nil
}

/*
	readInitial is the first method that should be invoked to create the frame
	instance based on the contents from a buffer. This method reads first 2
	bytes of a websocket frame which includes: fin (1 bit), rsv 1-3 (3 bits),
	opcode (4 bits), mask (1 bit) and payload length (7 bits). It accepts a
	buffer as an argument which will be used to read the frame from.
*/
func (f *frame) readInitial(b *bufio.Reader) error {
	// Read first 2 bytes.
	p, err := readFromBuffer(b, 2)

	if err != nil {
		return err
	}

	// Reading 'fin'
	if p[0]>>7 == 1 {
		f.fin = true
	}

	// Since library doesn't support extensions if RSV1-3 are non zeros, fail
	// connection
	if p[0]&112 /* 01110000 */ != 0 {
		return &CloseError{
			Code:   CloseProtocolError,
			Reason: "no support for extensions",
		}
	}

	// Reading 'opcode'
	f.opcode = int(p[0]) & 15 /* 00001111 */

	// if opcode doesn't exists, must stop connection
	if !opcodeExist(f.opcode) {
		return &CloseError{
			Code:   CloseProtocolError,
			Reason: fmt.Sprintf("unsupported opcode: %d", f.opcode),
		}
	}

	// Reading 'mask'
	if p[1]>>7 == 1 {
		f.masked = true
	}

	// Reading 'payload len'
	f.length = uint64(p[1]) & 127 /* 01111111 */

	return nil
}

/*
	readLength should be invoked after readInitial method and is used to read
	the next 2 (if f.length == 126) or 8 (if f.length == 127) bytes. If f.length
	is <= 125, no read operations are done to the buffer provided as an
	argument.
*/
func (f *frame) readLength(b *bufio.Reader) error {
	// If f.length is <= 125 it means that we already have the payload length,
	// thus stop read operation.
	if f.length <= 125 {
		return nil
	}

	// For when f.length == 126, read next 2 bytes.
	var l uint64 = 2

	// If f.length == 127, read next 8 bytes.
	if f.length == 127 {
		l = 8
	}

	// Read number of bytes based on f.length.
	u, err := readFromBuffer(b, l)

	if err != nil {
		return err
	}

	// Reset length
	f.length = 0

	// At this point the bytes that represent the real payload length has been
	// retrieved from the buffer. So the next thing to do is to convert the byte
	// slice (representing the length) to an integer by combining the bytes
	// together.
	//
	// Example: Let say the slice of bytes repesenting the payload length is
	// [134, 129] (or [10000110, 10000001] in binary).
	//
	// loop 1: f.length == 0
	// 		line 1: Bitwise left shift of 8
	// 			length = 0
	// 		line 2: Add the byte being traversed to f.length.
	// 			length = 1310000110
	//
	// loop 2: f.length == 134 (or 10000110)
	// 		line 1: Bitwise left shift of 8
	// 			length = 10000110 00000000 (i.e. 34304)
	// 		line 2: Add the byte being traversed to f.length.
	// 			length = 10000110 10000001 (i.e. 34433)
	for _, v := range u {
		f.length = f.length << 8
		f.length += uint64(v)
	}

	// Most Significant Bit must be 0.
	f.length = f.length & 9223372036854775807

	return nil
}

/*
	readMaskKey should be invoked after readLength method and is used to read
	the next 4 bytes from the buffer to retrieve the masking key. Note that if
	the payload data is not masked (f.masked == false - info retrieved from
	readInitial) no read operations are done to the buffer provided as an
	argument.
*/
func (f *frame) readMaskKey(b *bufio.Reader) error {
	// If payload is not masked, stop process
	if !f.masked {
		return nil
	}

	// Read 4 bytes for masking key
	p, err := readFromBuffer(b, 4)

	if err != nil {
		return err
	}

	// Store key in frame instance
	f.key = p

	return nil
}

/*
	readPayload should be invoked after readMaskKey method and is used to read
	the payload data from the buffer. The number of bytes to read are known from
	f.length (info retrieved from either readInitial or readLength). In addition
	to this if the payload data is masked (f.masked == true - info retrieved
	from readInitial) the payload data will also be decoded using the masking
	key provided with the frame (f.key - info retrieved from readMaskKey).
*/
func (f *frame) readPayload(b *bufio.Reader) error {
	// Read f.length bytes
	p, err := readFromBuffer(b, f.length)

	if err != nil {
		return err
	}

	if f.masked {
		// Unmask (decode) payload data
		mask(p, f.key)
	}

	// Store payload in frame instance.
	f.payload = p

	return nil
}

/*
	toBytes returns a representation of the frame instance as a slice of bytes.
	This method does not consider the values assigned to f.length and f.masked
	since these are calculated using the length of f.payload and value of f.key
	respectively.
*/
func (f *frame) toBytes() ([]byte, error) {
	if err := f.validate(); err != nil {
		return nil, err
	}

	// Slice of bytes used to contain the payload data.
	p := make([]byte, 2)

	// Include info for FIN bit.
	f.toBytesFin(p)

	// Include info for OPCODE bits.
	f.toBytesOpcode(p)

	// Include info for MASK bit.
	f.toBytesMasked(p)

	// Include info for PAYLOAD LEN bits.
	f.toBytesPayloadLength(p)

	// Append (if any) info for PAYLOAD LENGTH EXTENDED bits.
	p = append(p, f.toBytesPayloadLengthExt()...)

	// Append (if any) MASK KEY bits.
	p = append(p, f.key...)

	// Append (Masked) Payload data. bits
	p = append(p, f.toBytesPayloadData()...)

	// Append and PAYLOAD DATA bits and return whole payload
	return p, nil
}

/*
	validate verifies that the data of the frame instance will result in a valid
	websocket data frame.
*/
func (f *frame) validate() *CloseError {
	switch {
	// Opcode must exists.
	case !opcodeExist(f.opcode):
		{
			return &CloseError{
				Code:   CloseProtocolError,
				Reason: fmt.Sprintf("unsupported opcode: %d", f.opcode),
			}
		}
	// Masking key must have a valid length.
	case !validateKey(f.key):
		{
			return &CloseError{
				Code:   CloseProtocolError,
				Reason: "masking key must either be 0 or 4 bytes long",
			}
		}
	// Payload data must have a valid length.
	case !validatePayload(f.payload):
		{
			return &CloseError{
				Code:   CloseMessageTooBig,
				Reason: "maximum payload data exceeded",
			}
		}
	}
	return nil
}

/*
	toBytesFin is used by toBytes to include info in 'p' about the FIN bit of
	the frame instance. Note that this method should be invoked before
	toBytesOpcode method.
*/
func (f *frame) toBytesFin(p []byte) {
	if f.fin {
		p[0] = 128
	}
}

/*
	toBytesOpcode is used by toBytes to include info in 'p' about the OPCODE
	bits of the frame instance. Note that this method should be invoked after
	toBytesFin.
*/
func (f *frame) toBytesOpcode(p []byte) {
	p[0] += byte(f.opcode)
}

/*
	toBytesMasked is used by toBytes to include info in 'p' about the MASK bit
	of the frame instance. This method does not consider f.masked but instead it
	calculates the MASK bit value based on f.key. Note that this method should
	be invoked before toBytesPayloadLength.
*/
func (f *frame) toBytesMasked(p []byte) {
	if len(f.key) != 0 {
		p[1] = 128
	}
}

/*
	toBytesPayloadLength is used by toBytes to include info in 'p' about the
	PAYLOAD LENGTH bits of the frame instance. This method does not consider
	f.length but instead it calculates the PAYLOAD LENGTH value based on the
	payload that will be sent (f.payload). Note that this method should
	be invoked after toBytesMasked.
*/
func (f *frame) toBytesPayloadLength(p []byte) {
	l := len(f.payload)

	switch {
	case l <= 125:
		{
			p[1] += byte(l)
			return
		}
	case l <= 65535:
		{
			p[1] += 126
		}
	case l <= 9223372036854775807:
		{
			p[1] += 127
		}
	}
}

/*
	toBytesPayloadLengthExt is used by toBytes to include info about the PAYLOAD
	LENGTH EXTENDED bits. Just like toBytesPayloadLength, this method does not
	consider f.length but instead it calculates the PAYLOAD LENGTH EXTENDED bits
	using the payload that will be sent (f.payload).
*/
func (f *frame) toBytesPayloadLengthExt() []byte {
	l := len(f.payload)

	// If <= 125, stop process since the true length is already known.
	if l <= 125 {
		return nil
	}

	var p []byte

	switch {
	case l <= 65535:
		{
			// Convert to binary.
			p = make([]byte, 2)
			binary.BigEndian.PutUint16(p, uint16(l))
		}
	case l <= 9223372036854775807:
		{
			// Convert to binary.
			p = make([]byte, 8)
			binary.BigEndian.PutUint64(p, uint64(l))
		}
	}

	return p
}

/*
	toBytesPayloadData is used by toBytes to include info about the PAYLOAD
	DATA. This method also handles the masking of the payload data (f.payload).
	Note that just like toBytesMasked, this method does not consider f.masked
	but instead it directly checks for the masking key (f.key).
*/
func (f *frame) toBytesPayloadData() []byte {
	// Put payload into another slice of bytes - so that the payload in the
	// frame instance is left untouched.
	p := append([]byte{}, f.payload...)

	// If masking key is present, use it to mask the payload data.
	if len(f.key) == 4 {
		mask(p, f.key)
	}

	return p
}
