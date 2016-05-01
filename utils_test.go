package websocket

import (
	"bufio"
	"math/rand"
	"strings"
	"testing"
)

func TestMakeAcceptKey(t *testing.T) {
	e := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	k := makeAcceptKey("dGhlIHNhbXBsZSBub25jZQ==")
	if k != e {
		t.Errorf(`expected "%s" instead "%s" was returned.`, e, k)
	}
}

type payloadMock struct {
	p []byte
}

func (m *payloadMock) Read(p []byte) (int, error) {
	n := 0

	for i := range m.p {
		if i == len(p) {
			break
		}

		p[i] = m.p[i]
		n++
	}

	m.p = append(make([]byte, 0), m.p[n:]...)

	return n, nil
}

func newBuffer(d []byte) *bufio.Reader {
	p := &payloadMock{p: d}
	r := bufio.NewReader(p)
	return r
}

func TestReadFromBufferSingleRead(t *testing.T) {
	var c uint64 = 3
	p := []byte{120, 123, 54, 32, 102}
	b := newBuffer(p)

	n, err := readFromBuffer(b, c)

	if err != nil {
		t.Fatal("An unexpected error was returned while invoking readFromBuffer():", err)
	}

	if uint64(len(n)) != c {
		t.Errorf("Expected slice of bytes returned from readFromBuffer to be of the length '%d'. Instead it is '%d'.", c, len(n))
	}

	for i, v := range n {
		if v != p[i] {
			t.Fatalf("Expected slice of bytes to be '%v'. Instead it is '%v'.", p[:c], n)
		}
	}
}

func TestReadFromBufferMultiRead(t *testing.T) {
	// The slice to be read from the buffer must be greater than 4096. Since
	// this is the default size of a bufio buffer.
	// GO Ref: https://golang.org/src/bufio/bufio.go#L18
	p := make([]byte, 4100)

	for i := range p {
		rand.Seed(int64(i))
		p[i] = byte(rand.Intn(255))
	}

	b := newBuffer(p)

	readFromBuffer(b, 4090)
	n, err := readFromBuffer(b, 10)

	if err != nil {
		t.Error("Unexpected error was returned while invoking readFromBuffer:", err)
	}

	for i, v := range n {
		if v != p[i+4090] {
			t.Errorf("%v != %v", p[i+4090], v)
		}
	}
}

func TestStringExists(t *testing.T) {
	l := []string{"one", "two", "three"}

	type testCase struct {
		k string
		v int
	}

	testCases := []testCase{
		{k: "one", v: 0},
		{k: "four", v: -1},
	}

	for i, c := range testCases {
		r := stringExists(l, c.k)

		if r != c.v {
			t.Errorf(`Test Case %d: Expected stringExists("%s") to return '%d' instead returned '%d'`, i, c.k, c.v, r)
		}
	}
}

func TestHeaderToSlice(t *testing.T) {
	l := []string{"  both  ", "  left", "right  ", "none"}

	r := headerToSlice(strings.Join(l, ","))

	if len(l) != len(r) {
		t.Errorf("The length of the list of header value are not the same. '%d' != '%d'.", len(l), len(r))
	}

	if r[0] != "both" {
		t.Errorf(`Expected "both" instead got "%s".`, r[0])
	}

	if r[1] != "left" {
		t.Errorf(`Expected "left" instead got "%s".`, r[1])
	}

	if r[2] != "right" {
		t.Errorf(`Expected "right" instead got "%s".`, r[2])
	}

	if r[3] != "none" {
		t.Errorf(`Expected "none" instead got "%s".`, r[3])
	}
}

func TestRandomByteSlice(t *testing.T) {
	type testCase struct {
		l int
	}

	testCases := []testCase{
		{l: 2},
		{l: 6},
	}

	for i, c := range testCases {
		if b := randomByteSlice(c.l); len(b) != c.l*4 {
			t.Errorf("test case %d: expected slice of bytes to be '%d' in length, but it is '%d'", i, c.l*4, len(b))
		}
	}
}

func TestCloseErrorExist(t *testing.T) {
	type testCase struct {
		e int
		v bool
	}

	testCases := []testCase{
		// Should return false when opcode is invalid
		{e: 15, v: false},
		// Should return true when opcode is valid.
		{e: CloseNormalClosure, v: true},
	}

	for i, c := range testCases {
		if v := closeErrorExist(c.e); v != c.v {
			t.Errorf("test case %d: expected '%t' for '%d'", i, c.v, c.e)
		}
	}
}
