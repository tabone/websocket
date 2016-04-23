package websocket

import (
	"bufio"
	"testing"
)

func TestReadInitialForFin(t *testing.T) {
	type testCase struct {
		b *bufio.Reader
		v bool
	}

	testCases := []testCase{
		// When fin bit is '0' should set fin to false.
		testCase{b: newBuffer([]byte{1 /* 00000001 */, 0}), v: false},
		// When fin bit is '1' should set fin to true.
		testCase{b: newBuffer([]byte{129 /* 10000001 */, 0}), v: true},
	}

	for i, c := range testCases {
		f := &frame{}

		if err := f.readInitial(c.b); err != nil {
			t.Errorf("test case %d: unexpected error returned: %v", i, err)
		}

		if f.fin != c.v {
			t.Errorf("test case %d: expected 'fin' to be '%t'", i, c.v)
		}
	}
}

func TestReadInitialForOpcode(t *testing.T) {
	type testCase struct {
		b *bufio.Reader
		v int
	}

	// When opcode is valid, should not return an error.
	testCases := []testCase{
		// Without mask bit.
		testCase{b: newBuffer([]byte{0, 0}), v: OpcodeContinuation},
		testCase{b: newBuffer([]byte{1, 0}), v: OpcodeText},
		testCase{b: newBuffer([]byte{2, 0}), v: OpcodeBinary},
		testCase{b: newBuffer([]byte{8, 0}), v: OpcodeClose},
		testCase{b: newBuffer([]byte{9, 0}), v: OpcodePing},
		testCase{b: newBuffer([]byte{10, 0}), v: OpcodePong},

		// With mask bit.
		testCase{b: newBuffer([]byte{128, 0}), v: OpcodeContinuation},
		testCase{b: newBuffer([]byte{129, 0}), v: OpcodeText},
		testCase{b: newBuffer([]byte{130, 0}), v: OpcodeBinary},
		testCase{b: newBuffer([]byte{136, 0}), v: OpcodeClose},
		testCase{b: newBuffer([]byte{137, 0}), v: OpcodePing},
		testCase{b: newBuffer([]byte{138, 0}), v: OpcodePong},
	}

	for i, c := range testCases {
		f := &frame{}

		if err := f.readInitial(c.b); err != nil {
			t.Errorf("test case %d: unexpected error returned: %v", i, err)
		}

		if f.opcode != c.v {
			t.Errorf("test case %d: expected 'opcode' to be '%d'", i, c.v)
		}
	}
}

func TestReadInitialForRSVError(t *testing.T) {
	type testCase struct {
		b *bufio.Reader
	}

	// Library doesn't support extensions thus when extension bits are used,
	// lib should return an error.
	testCases := []testCase{
		testCase{b: newBuffer([]byte{17 /* 00010001 */, 0})},
		testCase{b: newBuffer([]byte{33 /* 00100001 */, 0})},
		testCase{b: newBuffer([]byte{49 /* 00110001 */, 0})},
		testCase{b: newBuffer([]byte{65 /* 01000001 */, 0})},
		testCase{b: newBuffer([]byte{81 /* 01010001 */, 0})},
		testCase{b: newBuffer([]byte{97 /* 01100001 */, 0})},
		testCase{b: newBuffer([]byte{113 /* 01110001 */, 0})},
	}

	for i, c := range testCases {
		f := &frame{}

		err := f.readInitial(c.b)

		if err == nil {
			t.Errorf("test case %d: an error was expected.", i)
		}

		e, k := err.(*CloseError)

		if !k {
			t.Errorf("test case %d: expected error to be of type '*CloseError' but it is '%T'.", i, e)
		}

		if e.Reason != "no support for extensions" {
			t.Errorf(`test case %d: expected error to have reason "no support for extensions", instead it got "%s".`, i, e.Reason)
		}
	}
}

// Should return an error if opcode is invalid.
func TestReadInitialForOpcodeError(t *testing.T) {
	f := &frame{}
	b := newBuffer([]byte{15, 0})

	err := f.readInitial(b)

	if err == nil {
		t.Error("unexpected error returned")
	}

	e, k := err.(*CloseError)

	if !k {
		t.Fatalf("expected error to be of type '*websocket.CloseError', but it is '%T'.", e)
	}

	if e.Reason != "unsupported opcode: 15" {
		t.Errorf(`expected error to have reason "unsupported opcode: 15", but it got "%s".`, e.Reason)
	}
}

func TestReadInitialForMasked(t *testing.T) {
	type testCase struct {
		b *bufio.Reader
		v bool
	}

	testCases := []testCase{
		// When masked bit is '0' should set masked to false.
		testCase{b: newBuffer([]byte{1, 0}), v: false},
		// When masked bit is '1' should set masked to true.
		testCase{b: newBuffer([]byte{1, 128}), v: true},
	}

	for i, c := range testCases {
		f := &frame{}

		if err := f.readInitial(c.b); err != nil {
			t.Errorf("test case %d: unexpected error returned: %v", i, err)
		}

		if f.masked != c.v {
			t.Errorf("test case %d: expected 'masked' to be '%t'", i, c.v)
		}
	}
}

func TestReadInitialForLength(t *testing.T) {
	type testCase struct {
		b *bufio.Reader
		v uint64
	}

	testCases := []testCase{
		// Should set length to 124 when payload len is 124.
		testCase{b: newBuffer([]byte{1, 124}), v: 124},
		testCase{b: newBuffer([]byte{1, 252}), v: 124},
		// Should set length to 125 when payload len is 125.
		testCase{b: newBuffer([]byte{1, 125}), v: 125},
		testCase{b: newBuffer([]byte{1, 253}), v: 125},
		// Should set length to 126 when payload len is 126.
		testCase{b: newBuffer([]byte{1, 126}), v: 126},
		testCase{b: newBuffer([]byte{1, 254}), v: 126},
		// Should set length to 127 when payload len is 127.
		testCase{b: newBuffer([]byte{1, 127}), v: 127},
		testCase{b: newBuffer([]byte{1, 255}), v: 127},
	}

	for i, c := range testCases {
		f := &frame{}

		if err := f.readInitial(c.b); err != nil {
			t.Errorf("test case %d: unexpected error returned: %v", i, err)
		}

		if f.length != c.v {
			t.Errorf("test case %d: expected 'length' to be '%d'", i, c.v)
		}
	}
}

func TestReadMaskKey(t *testing.T) {
	f := &frame{}
	p := []byte{102, 100, 1, 54}
	b := newBuffer(p)

	// When f.masked is false, it means that the payload is not masked and
	// therefore no key has been sent. For this reason f.key should be left
	// untouched.
	f.masked = false
	if err := f.readMaskKey(b); err != nil {
		t.Error("unexpected error returned:", err)
	}

	if len(f.key) != 0 {
		t.Error("expected f.key to be empty but it is:", len(f.key))
	}

	// When f.masked is true, it means that the payload is masked and therefore
	// the key must be read and stored in f.key.
	f.masked = true
	f.key = nil
	if err := f.readMaskKey(b); err != nil {
		t.Error("unexpected error returned:", err)
	}

	if len(f.key) != 4 {
		t.Errorf("expected f.key to be '4 bytes' long but it is '%d bytes'", len(f.key))
	}

	for i, v := range p {
		if v != f.key[i] {
			t.Fatalf("expected mask key to be '%v' but it is '%v'", p, f.key)
		}
	}
}

func TestReadLength(t *testing.T) {
	f := &frame{}

	type testCase struct {
		// initial length
		i uint64
		// final length
		l uint64
	}

	testCases := []testCase{
		testCase{i: 124, l: 124},
		testCase{i: 125, l: 125},
		testCase{i: 126, l: 65535},
		testCase{i: 127, l: 9223372036854775807},
	}

	for i, c := range testCases {
		f.length = c.i

		b := newBuffer([]byte{255, 255, 255, 255, 255, 255, 255, 255})
		if err := f.readLength(b); err != nil {
			t.Errorf("test case %d: unexpected error returned: %v", i, err)
		}

		if f.length != c.l {
			t.Errorf("test case %d: expected f.length to be '%d', but it is '%d'", i, c.l, f.length)
		}
	}
}

func TestReadPayload(t *testing.T) {
	type testCase struct {
		// Masked or not
		m bool
	}

	testCases := []testCase{
		testCase{m: false},
		testCase{m: true},
	}

	for i, c := range testCases {
		// Data Frame Recieved
		p := []byte{120, 15, 17}
		b := newBuffer(p)

		// Creation and config of frame instance.
		f := &frame{}
		f.key = []byte{10, 15, 1, 120}
		f.length = 2

		// Setting Masked.
		f.masked = c.m

		if err := f.readPayload(b); err != nil {
			t.Fatalf("test case %d: unexpected error was returned: %v", i, err)
		}

		// If masked unmask it.
		if f.masked {
			mask(f.payload, f.key)
		}

		if uint64(len(f.payload)) != f.length {
			t.Errorf("test case %d: expected length of f.payload to be '%d', but it is '%d'", i, f.length, len(f.payload))
		}

		for i, v := range f.payload {
			if v != p[i] {
				t.Fatalf("test case %d: expected slice of bytes to be '%v', but it is '%v'.", i, p[:f.length], f.payload)
			}
		}
	}
}

func TestToBytesFin(t *testing.T) {
	type testCase struct {
		v bool
		r byte
	}

	testCases := []testCase{
		// When f.fin is false first byte should not be affected
		testCase{v: false, r: 0},
		// When f.fin is true first byte should has its MSB == 1.
		testCase{v: true, r: 128},
	}

	for i, c := range testCases {
		f := &frame{fin: c.v}
		p := make([]byte, 1)

		f.toBytesFin(p)

		if p[0] != c.r {
			t.Errorf("test case %d: expected slice of bytes to be [%d] but it is [%d]", i, c.r, p[0])
		}
	}
}

func TestToBytesOpcode(t *testing.T) {
	type testCase struct {
		// Fin Value
		v bool
		// Opcode Value
		o int
		// Resultant byte
		r byte
	}

	testCases := []testCase{
		// With Fin == false
		testCase{v: false, o: OpcodeText, r: byte(OpcodeText)},
		// With Fin == true
		testCase{v: true, o: OpcodeText, r: 128 + byte(OpcodeText)},
	}

	for i, c := range testCases {
		f := &frame{fin: c.v, opcode: c.o}
		p := make([]byte, 1)

		f.toBytesFin(p)
		f.toBytesOpcode(p)

		if p[0] != c.r {
			t.Errorf("test case %d: expected slice of bytes to be [%d] but it is [%d]", i, c.r, p[0])
		}
	}
}

func TestToBytesMasked(t *testing.T) {
	type testCase struct {
		// Value of f.key.
		v []byte
		// Resultant byte.
		r byte
	}

	testCases := []testCase{
		testCase{v: nil, r: 0},
		testCase{v: []byte{1, 2, 3, 4}, r: 128},
	}

	for i, c := range testCases {
		f := frame{key: c.v}
		p := make([]byte, 2)

		f.toBytesMasked(p)

		if p[1] != c.r {
			t.Errorf("test case %d: expected slice of bytes to be [%d] but it is [%d]", i, c.r, p[1])
		}
	}
}

func TestToBytesPayloadLength(t *testing.T) {
	type testCase struct {
		m bool
		r byte
		l int
	}

	testCases := []testCase{
		// With Mask Bit (f.masked) set to false
		testCase{m: false, r: 124, l: 124},
		testCase{m: false, r: 125, l: 125},
		testCase{m: false, r: 126, l: 30000},
		testCase{m: false, r: 126, l: 65535},
		testCase{m: false, r: 127, l: 700000},
		// With Mask Bit (f.masked) set to true
		testCase{m: true, r: 128 + 124, l: 124},
		testCase{m: true, r: 128 + 125, l: 125},
		testCase{m: true, r: 128 + 126, l: 30000},
		testCase{m: true, r: 128 + 126, l: 65535},
		testCase{m: true, r: 128 + 127, l: 700000},
		// testCase{m: false, r: 127, l: 9223372036854775807},
	}

	for i, c := range testCases {
		p := make([]byte, 2)
		f := frame{payload: make([]byte, c.l)}

		if c.m {
			f.key = []byte{1, 2, 3, 4}
		}

		f.toBytesMasked(p)
		f.toBytesPayloadLength(p)

		if p[1] != c.r {
			t.Errorf("test case %d: expected slice of bytes to be [%d] but it is [%d]", i, c.r, p[1])
		}
	}
}

func TestToBytesPayloadLengthExt(t *testing.T) {
	type testCase struct {
		l int
		r []byte
	}

	testCases := []testCase{
		// Length Known.
		testCase{l: 124, r: nil},
		// Length Known.
		testCase{l: 125, r: nil},
		// Read next 2 bytes.
		testCase{l: 30000, r: []byte{117, 48}},
		// Read next 2 bytes.
		testCase{l: 65535, r: []byte{255, 255}},
		// Read next 8 bytes.
		testCase{l: 700000, r: []byte{0, 0, 0, 0, 0, 10, 174, 96}},
	}

	for i, c := range testCases {
		f := frame{payload: make([]byte, c.l)}
		p := f.toBytesPayloadLengthExt()

		if len(p) != len(c.r) {
			t.Errorf("test case %d: expected length to be '%d' but it is '%d'", i, len(c.r), len(p))
		}

		for ci, cv := range c.r {
			if cv != p[ci] {
				t.Errorf("test case %d: Expected slice of bytes to be %v but it is %v", i, c.r, p)
				break
			}
		}
	}
}

func TestToBytesPayloadData(t *testing.T) {
	type testCase struct {
		m []byte
		p []byte
	}

	testCases := []testCase{
		// When masking key is present and valid, payload must be masked.
		testCase{p: []byte{3, 4, 5, 6}, m: nil},
		// When masking key is not present, payload must not be masked.
		testCase{p: []byte{3, 4, 5, 6}, m: []byte{1, 2, 3, 4}},
	}

	for i, c := range testCases {
		f := &frame{key: c.m, payload: c.p}

		p := f.toBytesPayloadData()

		if c.m != nil {
			mask(p, c.m)
		}

		for ci, cv := range c.p {
			if cv != p[ci] {
				t.Errorf("test case %d: Expected slice of bytes to be %v but it is %v", i, c.p, p)
			}
		}
	}
}
