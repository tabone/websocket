package websocket

import (
	"net/http"
	"sync"
)

// wsVersion is the websocket version this library supports.
const wsVersion = "13"

// Request represents the HTTP Request that will be upgraded to the WebSocket
// protocol once it is validated.
type Request struct {
	/*
		request is the http request to be upgraded.
	*/
	request *http.Request

	/*
		CheckOrigin is the function which will be used to validate the ORIGIN
		HTTP Header of the request. By default this method will fail the opening
		handshake when the origin is not the same. This method can be overridden
		during the initiation of the Request struct.
	*/
	CheckOrigin func(r *http.Request) bool

	/*
		SubProtocol name which the server has agreed to use from the list
		provided by the client (through the Sec-WebSocket-Protocol HTTP Header
		Field). Before sending the servers opening handshake response, checks
		are made to verify that the chosen protocol was indeed been provided as
		an option from the client. If this is not the case, the HTTP
		Sec-WebSocket-Protocol HTTP Response Header Field is not sent
	*/
	SubProtocol string
}

// Upgrade is used to upgrade the HTTP connection to use the WS protocol once
// the client request is validated.
func (q *Request) Upgrade(w http.ResponseWriter, r *http.Request) (*Socket, error) {
	// Store a reference to the HTTP Request.
	q.request = r

	// Check origin.
	// Ref spec: https://tools.ietf.org/html/rfc6455#section-4.2.2
	if err := q.handleOrigin(); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil, err
	}

	// Check websocket version.
	// Ref spec: https://tools.ietf.org/html/rfc6455#section-4.2.2
	if err := validateWSVersionHeader(r); err != nil {
		w.Header().Set("Sec-WebSocket-Version", wsVersion)
		http.Error(w, "Upgrade Required", 426)
		return nil, err
	}

	// Check handshake request.
	// Ref spec: https://tools.ietf.org/html/rfc6455#section-4.2.2
	if err := validateRequest(r); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return nil, err
	}

	// At this point, the clients handshake request is valid and therefore the
	// connection can be upgraded to use the ws protocol.
	s, err := q.upgrade(w)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}

	return s, nil
}

func (q *Request) upgrade(w http.ResponseWriter) (*Socket, error) {
	// Take control of the net.Conn instance.
	h, k := w.(http.Hijacker)

	if !k {
		return nil, &OpenError{Reason: "assertion failed with current http.ResponseWriter instance"}
	}

	conn, buf, err := h.Hijack()
	if err != nil {
		return nil, err
	}

	// Build the HTTP Header response code required for the ws opening
	// handshake.
	// From RFC2616: https://www.w3.org/Protocols/rfc2616/rfc2616-sec6.html
	resp := "HTTP/1.1 101 Switching Protocols\n"
	resp += "Upgrade: websocket\n"
	resp += "Connection: upgrade\n"
	resp += "Sec-WebSocket-Version: " + wsVersion + "\n"

	// If server has agreed to use a sub-protocol, the chosen sub-protocol needs
	// to be an option provided by the clients endpoint. If not, the
	// Sec-WebSocket-Protocol HTTP Header field is not sent.
	if q.SubProtocol != "" && stringExists(q.ClientSubProtocols(), q.SubProtocol) != -1 {
		resp += "Sec-WebSocket-Protocol: " + q.SubProtocol + "\n"
	}

	// Generate the accept key based on the challenge key provided by the
	// client and include it inside 'Sec-WebSocket-Accept' response header
	// field.
	acceptKey := makeAcceptKey(q.request.Header.Get("Sec-WebSocket-Key"))
	resp += "Sec-WebSocket-Accept: " + acceptKey + "\n\n"

	// Send response
	buf.WriteString(resp)
	buf.Flush()

	// Create and return socket.
	return &Socket{
		conn:       conn,
		buf:        buf,
		server:     true,
		writeMutex: &sync.Mutex{},
	}, nil
}

// handleOrigin is used to invoke either the CheckOrigin method provided by the
// user or the default method (if the user doesn't provide one).
func (q *Request) handleOrigin() *OpenError {
	fn := q.CheckOrigin

	if fn == nil {
		fn = checkOrigin
	}

	if !fn(q.request) {
		return &OpenError{Reason: `failure due to origin.`}
	}

	return nil
}

// ClientSubProtocols returns the list of Sub Protocols the client can interact
// with.
//
// From spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
func (q *Request) ClientSubProtocols() []string {
	return headerToSlice(q.request.Header.Get("Sec-WebSocket-Protocol"))
}

// ClientExtensions returns the list of Extensions the client can interact
// with.
//
// From spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
func (q *Request) ClientExtensions() []string {
	return headerToSlice(q.request.Header.Get("Sec-WebSocket-Extensions"))
}
