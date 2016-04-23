package websocket

import (
	"bufio"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

/*
	ErrSocketClosed is the error returned when a user tries to send a frame with
	a closed socket.
*/
var ErrSocketClosed = errors.New("socket has been closed")

/*
	WebSocket Error codes.
	Ref Spec: https://tools.ietf.org/html/rfc6455#section-7.4.1
*/
const (
	CloseNormalClosure           int = 1000
	CloseGoingAway               int = 1001
	CloseProtocolError           int = 1002
	CloseUnsupportedData         int = 1003
	CloseNoStatusReceived        int = 1005
	CloseAbnormalClosure         int = 1006
	CloseInvalidFramePayloadData int = 1007
	ClosePolicyViolation         int = 1008
	CloseMessageTooBig           int = 1009
	CloseMandatoryExtension      int = 1010
	CloseInternalServerErr       int = 1011
	CloseTLSHandshake            int = 1015
)

/*
	Represents the state of the Socket instance
*/
const (
	/*
		stateOpened will be the state when the socket instance is open.
	*/
	stateOpened int = 0

	/*
		stateClosing will be the state when the socket instance is in the middle
		of the closing handshake.
	*/
	stateClosing int = 1

	/*
		stateClosed will be the state when the socket instance is closed.
	*/
	stateClosed int = 2
)

/*
	Socket represents a socket endpoint.
*/
type Socket struct {
	/*
		conn is the underlying tcp connection.
	*/
	conn net.Conn

	/*
		buf is a buffered version of the underlying tcp connection.
	*/
	buf *bufio.ReadWriter

	/*
		server indicates whether the socket instance represents a server or a
		client endpoint.
	*/
	server bool

	/*
		state is the current state of the socket instance.
	*/
	state int

	/*
		closeDelay is the duration the socket instance will wait until it closes
		the underlying tcp connection once the closing handshake has been
		completed.

		The websocket rfc suggests that when the closing handshake is completed
		the underlying tcp connection should first be terminated by the server
		endpoint. Having said this it doesn't restrict the client endpoint to do
		so itself. CloseDelay is the maximum time the socket instance will wait
		before it closes the tcp connection.

		Note: Server endpoints should always have this property set to 0.

		Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.5.1
	*/
	CloseDelay time.Duration

	/*
		readHandler is invoked whenever a text or binary frame is recieved. The
		opcode and payload data are provided as args respectively.
	*/
	ReadHandler func(int, []byte)

	/*
		pingHandler is invoked whenever a ping frame is recieved. The payload
		data is provided as arg.
	*/
	PingHandler func([]byte)

	/*
		pongHandler is invoked whenever a pong frame is recieved. The payload
		data is provided as arg.
	*/
	PongHandler func([]byte)

	/*
		closeHandler is invoked whenever the websocket connection is closed. The
		reason for the closure is provided as an arg.
	*/
	CloseHandler func(error)

	/*
		closeError contains the error which caused the websocket connection to
		terminate. This is then provided as an arg when invoking the close
		handler once the underlying tcp connection is terminated.
	*/
	closeError error

	/*
		writeMutex is used to queue the write functionality of a socket
		instance.
	*/
	writeMutex *sync.Mutex
}

/*
	Listen is used to start listening for new frames sent by the connected
	endpoint.
*/
func (s *Socket) Listen() {
	s.read()
}

func (s *Socket) read() {
Read:
	for {
		// Read frame
		f, err := newFrame(s.buf.Reader)

		if s.state == stateClosed {
			break Read
		}

		if err != nil {
			// If an error occured due to something which doesn't conform with
			// the websocket rfc, use the error itself as a reason.
			if c, k := err.(*CloseError); k {
				s.CloseWithError(c)
				return
			}

			// When EOF returns it means that the other endpoint isn't reachable
			// and thus there won't be the need to initate the closing
			// handshake.
			if err == io.EOF {
				s.closeError = &CloseError{
					Code:   CloseAbnormalClosure,
					Reason: "abnormal closure",
				}
				s.TCPClose()
				break Read
			}

			// When Read times out or connection is closed the other endpoing
			// won't be reachable and thus there won't be the need to initiate
			// the closing handshake.
			if _, k := err.(*net.OpError); k {
				s.closeError = &CloseError{
					Code:   CloseAbnormalClosure,
					Reason: "abnormal closure",
				}
				s.TCPClose()
				break Read
			}

			// Else use a generic error.
			s.CloseWithError(&CloseError{
				Code:   CloseProtocolError,
				Reason: "protocol error",
			})

			return
		}

		// If Socket instance represents a server endpoint, payload data must be
		// masked.
		// Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.1
		if s.server && !f.masked {
			s.CloseWithError(&CloseError{
				Code:   CloseProtocolError,
				Reason: "expected payload to be masked",
			})
			return
		}

		// If Socket instance represents a client endpoint, payload data must
		// not be masked.
		// Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.1
		if !s.server && f.masked {
			s.CloseWithError(&CloseError{
				Code:   CloseProtocolError,
				Reason: "expected payload to not be masked",
			})
			return
		}

		switch f.opcode {
		case OpcodeText, OpcodeBinary:
			{
				s.callReadHandler(f.opcode, f.payload)
			}
		case OpcodePing:
			{
				s.callPingHandler(f.payload)
			}
		case OpcodePong:
			{
				s.callPongHandler(f.payload)
			}
		case OpcodeClose:
			{
				// Create a new CloseError using the payload data
				c, cerr := NewCloseError(f.payload)

				// Store close error for close handler.
				s.closeError = c

				// If the state of the socket instance is CLOSING, it means that
				// the closing handshake has been initiated from this socket
				// instance and the retrieved frame was the acknowledge close
				// frame. At this point the closing handshake has been completed
				// and therefore the underlying tcp connection can be closed,
				// since the connected endpoint won't be waiting for furthur
				// frames.
				if s.state == stateClosing {
					// closing handshake has been finalized therefore close tcp
					// connection.
					s.tcpClose()
					// Stop reading from connection.
					break Read
				}

				// If the state of the socket instance is not CLOSING, it means
				// that the closing handshake has been initiated by the
				// connected endpoint and therefore it is still waiting for the
				// acknowledgement close frame.
				s.state = stateClosing

				// The acknowledgment close frame to be sent will echo the
				// status code of the close frame just recieved.
				// Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.5.1
				var b []byte

				// If the status code of the close frame recieved is valid, echo
				// it. Else leave the payload data of the acknowledgement close
				// frame empty.
				if cerr == nil {
					b = c.toBytesCode()
				}

				// Send acknowledgement close frame.
				s.Write(OpcodeClose, b)

				// At this point the closing handshake would have been finalized
				// therefore the tcp connection can be closed.
				s.tcpClose()

				// Stop reading from connection.
				break Read
			}
		}
	}
}

/*
	Write is used to send new data frames to the connected endpoint. It accepts
	two arguments 'o' opcode, 'p' payload data.
*/
func (s *Socket) Write(o int, p []byte) error {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	// Before writing make sure that the socket instance is still in an open
	// state.
	if s.state == stateClosed {
		return ErrSocketClosed
	}

	// Create a frame instance which will represent the frame to be sent.
	f := &frame{
		fin:     true,
		opcode:  o,
		payload: p,
	}

	// If the socket instance represents a client endpoint, the payload data
	// must be masked.
	if !s.server {
		// Generate random mask key
		f.key = randomByteSlice(1)
	}

	// Get a []byte representation of the frame instance.
	b, err := f.toBytes()

	// If an error is not nil, since the error doesn't relate with the socket
	// connection itself, the error is returned.
	if err != nil {
		return err
	}

	// Send frame
	s.buf.Write(b)
	if err := s.buf.Flush(); err != nil {
		// Store error.
		s.closeError = err

		// Close TCP Connection.
		s.TCPClose()

		// Since the error is related with the socket connection the error is
		// not returned but passed to the close handler.
		return nil
	}

	// If frame sent is a close frame, change state to closing.
	if f.opcode == OpcodeClose {
		s.state = stateClosing
	}

	return nil
}

/*
	SetReadDeadline sets the deadline for future Read calls. A zero value for t
	means Read will not time out.
*/
func (s *Socket) SetReadDeadline(t time.Time) {
	s.conn.SetReadDeadline(t)
}

/*
	SetWriteDeadline sets the deadline for future Write calls. Even if write
	times out, it may return n > 0, indicating that some of the data was
	successfully written. A zero value for t means Write will not time out.
*/
func (s *Socket) SetWriteDeadline(t time.Time) {
	s.conn.SetWriteDeadline(t)
}

/*
	callReadHandler invokes the read handler provided by the user (if any).
*/
func (s *Socket) callReadHandler(o int, p []byte) {
	if s.ReadHandler != nil {
		s.ReadHandler(o, p)
	}
}

/*
	callPingHandler first tries to invoke the ping handler provided by the
	user. If the user hasn't provided one it invokes the default functionality.
*/
func (s *Socket) callPingHandler(p []byte) {
	if s.PingHandler != nil {
		s.PingHandler(p)
		return
	}
	s.defaultPingHandler(p)
}

/*
	defaultPingHandler sends a pong frame with the same payload data of the ping
	frame just recieved.

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.5.3
*/
func (s *Socket) defaultPingHandler(p []byte) {
	s.Write(OpcodePong, p)
}

/*
	callPongHandler invokes the pong handler provided by the user (if any).
*/
func (s *Socket) callPongHandler(p []byte) {
	if s.PongHandler != nil {
		s.PongHandler(p)
		return
	}
}

/*
	callCloseHandler first tries to invoke the close handler provided by the
	user.
*/
func (s *Socket) callCloseHandler(e error) {
	if s.CloseHandler != nil {
		s.CloseHandler(e)
	}
}

/*
	TCPClose closes the underlying tcp connection if it hasn't already been
	closed.
*/
func (s *Socket) TCPClose() {
	// If socket has already been closed, don't reclose the tcp connection
	if s.state == stateClosed {
		return
	}

	// Change state of socket instance to closed.
	s.state = stateClosed

	// Close tcp connection
	s.conn.Close()

	// Invoke close handler.
	s.callCloseHandler(s.closeError)
}

/*
	tcpClose closes the underlying tcp connection after s.CloseDelay seconds if
	it hasn't already been closed . More info on why this is needed documented
	in s.CloseDelay.
*/
func (s *Socket) tcpClose() {
	// If socket has already been closed, don't reclose the tcp connection
	if s.state == stateClosed {
		return
	}

	if s.CloseDelay > 0 {
		t := time.NewTicker(time.Second * s.CloseDelay)
		<-t.C
	}

	// Close tcp connection
	s.TCPClose()
}

/*
	Close initiates the normal closures (1000) closing handshake.
*/
func (s *Socket) Close() {
	s.CloseWithError(&CloseError{
		Code:   CloseNormalClosure,
		Reason: "normal closure",
	})
}

/*
	CloseWithError initiates the closing handshake.
*/
func (s *Socket) CloseWithError(e *CloseError) {
	// Store error.
	s.closeError = e

	// Start the closing handshake
	b, _ := e.ToBytes()
	s.Write(OpcodeClose, b)
}
