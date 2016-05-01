package websocket

import (
	"bufio"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

// Dialer is a websocket client.
type Dialer struct {
	/*
		Header to be included in the opening handshake request.
	*/
	Header http.Header

	/*
		SubProtocols which the client supports.
	*/
	SubProtocols []string

	/*
		TLSConfig is used to configure the TLS client.
	*/
	TLSConfig *tls.Config
}

// Dial is the method used to start the websocket connection.
func (d *Dialer) Dial(u string) (*Socket, *http.Response, error) {
	// Parse URL to return a valid URL instance.
	l, err := parseURL(u)
	if err != nil {
		return nil, nil, err
	}

	// Get a valid websocket opening handshake request instance.
	q := d.createRequest(l)

	// Connect with the websocket server.
	// Ref Spec: https://tools.ietf.org/html/rfc6455#section-3
	conn, err := net.Dial("tcp", l.Host+"/"+l.Path+"?"+l.RawQuery)
	if err != nil {
		return nil, nil, err
	}

	// When the connection will be over TLS, we need to do the TLS handshake.
	if l.Scheme == "wss" {
		g := d.TLSConfig

		// Create tls config instance if user hasn't specified one since it is
		// required.
		if g == nil {
			g = &tls.Config{}
		}

		// If ServerName is empty, use the host provided by the user.
		if g.ServerName == "" {
			g.ServerName = strings.Split(l.Host, ":")[0]
		}

		// Change the current conenction to a secure one.
		c := tls.Client(conn, g)

		// Do the handshake.
		if err := c.Handshake(); err != nil {
			return nil, nil, err
		}

		conn = c
	}

	// Send request
	if err := q.Write(conn); err != nil {
		return nil, nil, err
	}

	// Buffer connection.
	b := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// Read response
	r, err := http.ReadResponse(b.Reader, q)

	if err != nil {
		return nil, nil, err
	}

	// Validate response.
	if err := validateResponse(r); err != nil {
		return nil, nil, err
	}

	return &Socket{
		conn:       conn,
		buf:        b,
		writeMutex: &sync.Mutex{},
	}, r, nil
}

// createOpeningHandshakeRequest is used to return a valid websocket opening
// handshake client request.
// 
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.1
func (d *Dialer) createRequest(l *url.URL) *http.Request {
	// Initialize header if not already initialized.
	if d.Header == nil {
		d.Header = make(http.Header)
	}

	// When using the default port the Host header field should only consist of
	// the host (no port is shown).
	t := l.Host

	switch l.Scheme {
	case "ws":
		{
			re := regexp.MustCompile(":22$")
			t = re.ReplaceAllString(t, "")
		}
	case "wss":
		{
			re := regexp.MustCompile(":443$")
			t = re.ReplaceAllString(t, "")
		}
	}

	// Include headers
	d.Header.Set("Host", t)
	d.Header.Set("Upgrade", "websocket")
	d.Header.Set("Connection", "upgrade")
	d.Header.Set("Sec-WebSocket-Version", "13")
	d.Header.Set("Sec-WebSocket-Key", makeChallengeKey())
	d.Header.Set("Sec-WebSocket-Protocol", strings.Join(d.SubProtocols, ", "))

	// Create request instance
	q := &http.Request{
		Method:     "GET",
		URL:        l,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     d.Header,
		Host:       l.Host,
	}

	return q
}
