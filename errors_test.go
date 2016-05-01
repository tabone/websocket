package websocket

import (
	"testing"
)

func TestCloseErrorToBytes(t *testing.T) {
	type testCase struct {
		c int
		r string
		b []byte
	}

	testCases := []testCase{
		testCase{c: 1001, r: "normal closure", b: []byte{3, 233, 110, 111, 114, 109, 97, 108, 32, 99, 108, 111, 115, 117, 114, 101}},
		testCase{c: 1001, r: "", b: []byte{3, 233}},
	}

	for i, c := range testCases {
		e := &CloseError{Code: c.c, Reason: c.r}

		b, err := e.ToBytes()

		if err != nil {
			t.Errorf(`test case %d: unexpected error`, i)
		}

		if len(b) != len(c.b) {
			t.Errorf(`test case %d: unexpected slice of bytes`, i)
		}

		same := true
		for bi, bv := range b {
			if bv != c.b[bi] {
				same = false
				break
			}
		}

		if !same {
			t.Errorf(`test case %d: unexpected slice of bytes`, i)
		}
	}
}

func TestCloseErrorToBytesError(t *testing.T) {
	b := []byte{3, 237, 110, 111, 32, 115, 116, 97, 116, 117, 115, 32, 114, 101, 99, 105, 101, 118, 101, 100}

	c := &CloseError{Code: 0, Reason: "woops"}
	e, err := c.ToBytes()

	if err == nil {
		t.Error("expected an error")
	}

	same := true
	for bi, bv := range b {
		if bv != e[bi] {
			same = false
			break
		}
	}

	if !same {
		t.Errorf(`unexpected slice of bytes`)
	}
}

func TestNewCloseError(t *testing.T) {
	type testCase struct {
		c int
		r string
		b []byte
	}

	testCases := []testCase{
		testCase{c: 1001, r: "normal closure", b: []byte{3, 233, 110, 111, 114, 109, 97, 108, 32, 99, 108, 111, 115, 117, 114, 101}},
		testCase{c: 1001, r: "", b: []byte{3, 233}},
	}

	for i, c := range testCases {
		e, err := NewCloseError(c.b)

		if err != nil {
			t.Errorf(`test case %d: unexpected error`, i)
		}

		if e.Code != c.c {
			t.Errorf("test case %d: expected Code to be '%d', but it is '%d'", i, c.c, e.Code)
		}

		if e.Reason != c.r {
			t.Errorf(`test case %d: expected Reason to be "%s", but it is "%s"`, i, c.r, e.Reason)
		}
	}
}

func TestNewCloseErrorError(t *testing.T) {
	type testCase struct {
		p []byte
	}

	testCases := []testCase{
		testCase{p: make([]byte, 0)},
		testCase{p: []byte{3, 133}},
	}

	for i, c := range testCases {
		c, err := NewCloseError(c.p)
		r := "no status recieved"

		if err == nil {
			t.Errorf("test case %d: expected an error", i)
		}

		if c.Code != CloseNoStatusReceived {
			t.Errorf("test case %d, expected Code to be '%d', but it is '%d'", i, CloseNoStatusReceived, c.Code)
		}

		if c.Reason != r {
			t.Errorf(`test case %d, expected Reason to be "%s", but it is "%s"`, i, r, c.Reason)
		}
	}
}
