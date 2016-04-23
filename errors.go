package websocket

import (
	"encoding/binary"
	"errors"
	"fmt"
)

/*
	CloseError represents errors related to the websocket closing handshake.
*/
type CloseError struct {
	Code   int
	Reason string
}

/*
	Error implements the built in error interface.
*/
func (c *CloseError) Error() string {
	return fmt.Sprintf("Close Error: %d %s", c.Code, c.Reason)
}

/*
	ToBytes returns the representation of a CloseError instance in a []bytes
	that conforms with the way the websocket rfc expects the payload data of
	CLOSE FRAMES to be.

	While generating the []bytes, if the CloseError instance has an invalid
	error code, it will instead create the representation of a 'No Status
	Recieved Error' (i.e. 1005).

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.5.1
*/
func (c *CloseError) ToBytes() ([]byte, error) {
	// Validate Error Code
	if !closeErrorExist(c.Code) {
		// If it is not valid, return bytes for No Status Recieved error.
		n := &CloseError{
			Code:   CloseNoStatusReceived,
			Reason: "no status recieved",
		}
		b, _ := n.ToBytes()
		return b, errors.New("invalid error code")
	}

	return append(c.toBytesCode(), []byte(c.Reason)...), nil
}

/*
	toBytesCode is used to get a representation of the CloseError instance
	status code in []bytes.
*/
func (c *CloseError) toBytesCode() []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(c.Code))
	return b
}

/*
	NewCloseError is used to create a new CloseError instance by parsing 'b'. In
	order for this to happen the []bytes needs to conform with the way the
	websocket rfc expects the payload data of CLOSE FRAMES to be.

	While parsing if the error code (i.e. first two bytes) is invalid, it will
	default the CloseError instance returned to represent a 'No Status Recieved
	Error' (i.e. 1005).

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.5.1
*/
func NewCloseError(b []byte) (*CloseError, error) {
	var c int

	if len(b) >= 2 {
		cb := b[:2]
		c = int(binary.BigEndian.Uint16(cb))
	}

	if !closeErrorExist(c) {
		return &CloseError{
			Code:   CloseNoStatusReceived,
			Reason: "no status recieved",
		}, errors.New("invalid error code")
	}

	return &CloseError{
		Code:   c,
		Reason: string(b[2:]),
	}, nil
}

/*
	OpenError represents errors related to the websocket opening handshake.
*/
type OpenError struct {
	Reason string
}

/*
	Error implements the built in error interface.
*/
func (h *OpenError) Error() string {
	return "Handshake Error: " + h.Reason
}
