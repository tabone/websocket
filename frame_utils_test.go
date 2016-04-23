package websocket

import "testing"

func TestOpcodeExist(t *testing.T) {
	type testCase struct {
		o int
		v bool
	}

	testCases := []testCase{
		// Should return false when opcode is invalid
		testCase{o: 15, v: false},
		// Should return true when opcode is valid.
		testCase{o: OpcodeText, v: true},
	}

	for i, c := range testCases {
		if v := opcodeExist(c.o); v != c.v {
			t.Errorf("test case %d: expected '%t' for '%d'", i, c.v, c.o)
		}
	}
}

func TestValidateKey(t *testing.T) {
	type testCase struct {
		k []byte
		r bool
	}

	testCases := []testCase{
		testCase{k: []byte{1, 2, 3, 4}, r: true},
		testCase{k: []byte{}, r: true},
		testCase{k: []byte{1, 2, 3, 4, 5}, r: false},
		testCase{k: []byte{1, 2, 3}, r: false},
	}

	for i, c := range testCases {
		if validateKey(c.k) != c.r {
			t.Errorf("test case %d: expected '%t' for %v", i, c.r, c.k)
		}
	}
}

func TestValidatePayload(t *testing.T) {
	type testCase struct {
		l uint64
		r bool
	}

	testCases := []testCase{
		testCase{l: 125, r: true},
		// testCase{l: 9223372036854775807, r: true},
		// testCase{l: 9223372036854775808, r: false},
	}

	for i, c := range testCases {
		b := make([]byte, c.l)
		if validatePayload(b) != c.r {
			t.Errorf("test case %d: expected '%t' for payload of size '%d'", i, c.r, c.l)
		}
	}
}
