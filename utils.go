package websocket

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"io"
	"math/rand"
	"strings"
	"time"
)

/*
	wsAcceptSalt is the GUID used by the WebSocket protocol to generate the
	value for the "Sec-Websocket-Accept" response HTTP Header field.
*/
const wsAcceptSalt string = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

/*
	makeAcceptKey is used to generate the Accept Key which is then sent to the
	client using the 'Sec-Websocket-Accept' Response Header Field. This is used
	to prevent an attacker from ticking the server.

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-1.3
*/
func makeAcceptKey(k string) string {
	h := sha1.New()
	io.WriteString(h, k+wsAcceptSalt)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

/*
	readFromBuffer reads from the buffer (b) provided the number of specified
	bytes (l).
*/
func readFromBuffer(b *bufio.Reader, l uint64) ([]byte, error) {
	p := make([]byte, l)

	// If the number of buffered bytes will accomodate the number of bytes
	// requested, read once and return the read bytes.
	if uint64(b.Buffered()) >= l {
		_, err := b.Read(p)
		return p, err
	}

	// If the user requires more bytes than there is buffered, the buffer will
	// be read from multiple times.

	// Total number of bytes read from buffer.
	n := 0

	for {
		// Temporary slice of bytes.
		t := make([]byte, l)

		// Read from buffer and put read bytes in temporary slice of bytes.
		i, err := b.Read(t)

		if err != nil {
			return nil, err
		}

		// Append bytes to the slice of bytes to be returned.
		p = append(p[:n], t[:i]...)

		// Increment the total number of bytes with the bytes read.
		n += i

		// If the total number of bytes is the same as the number of bytes
		// requested, stop read operation and read bytes.
		if uint64(n) == l {
			break
		}
	}

	return p, nil
}

/*
	stringExists is a utility function used to check whether a slice of string
	('l') contains a particular value ('k'). If it does, its position will be
	returned otherwise '-1' is returned.
*/
func stringExists(l []string, k string) int {
	for i, v := range l {
		if k == v {
			return i
		}
	}

	return -1
}

/*
	headerToSlice is used to turn the values of a multi value HTTP Header field
	to a slice of string.

	From RFC2616: https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html#sec4.2
*/
func headerToSlice(v string) []string {
	l := strings.Split(v, ",")

	for i, v := range l {
		l[i] = strings.Trim(v, " ")
	}

	return l
}

/*
	randomByteSlice is used to generate a byte slice of random 32 bit integers.
*/
func randomByteSlice(i int) []byte {
	// Slice of bytes which will grow to be 16 bytes in length once the
	// operation is ready. This slice will then be used to generate the key to
	// be sent with the clients opening handshake using the Sec-Websocket-Key
	// Header.
	b := make([]byte, 0)

	// Set seed.
	rand.Seed(time.Now().UnixNano())

	// The challenge key must be 16 bytes in length.
	for l := 0; l < i; l++ {
		// Temp slice
		t := make([]byte, 4)

		// Generate a random 32bit number and store its binary value in 't'.
		binary.BigEndian.PutUint32(t, rand.Uint32())

		// Finally append the random generated number to 'b'.
		b = append(b, t...)
	}

	return b
}

/*
	closeErrorExist returns whether the error number provided as an argument is
	a valid error number or not.
*/
func closeErrorExist(i int) bool {
	switch i {
	case CloseNormalClosure, CloseGoingAway, CloseProtocolError, CloseUnsupportedData, CloseNoStatusReceived, CloseAbnormalClosure, CloseInvalidFramePayloadData, ClosePolicyViolation, CloseMessageTooBig, CloseMandatoryExtension, CloseInternalServerErr, CloseTLSHandshake:
		{
			return true
		}
	}
	return false
}
