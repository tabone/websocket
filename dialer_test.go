package websocket

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestDialerCreateRequestNilHeader(t *testing.T) {
	d := &Dialer{Header: nil}

	q := d.createRequest(&url.URL{})

	if q.Header == nil {
		t.Errorf("expected header to be initialized")
	}
}

func TestDialerCreateRequestNonNilHeader(t *testing.T) {
	h := make(http.Header)
	k := "testKey"
	v := "testValue"

	h.Add(k, v)

	d := &Dialer{Header: h}

	q := d.createRequest(&url.URL{})

	if q.Header.Get(k) != v {
		t.Errorf("expected header to be the one provided in dialer instance")
	}
}

func TestDialerCreateRequestHostHeader(t *testing.T) {
	d := &Dialer{}

	type testCase struct {
		u *url.URL
		v string
	}

	testCases := []testCase{
		testCase{u: &url.URL{Scheme: "ws", Host: "localhost:22"}, v: "localhost"},
		testCase{u: &url.URL{Scheme: "wss", Host: "localhost:443"}, v: "localhost"},
		testCase{u: &url.URL{Scheme: "ws", Host: "localhost:80"}, v: "localhost:80"},
		testCase{u: &url.URL{Scheme: "wss", Host: "localhost:80"}, v: "localhost:80"},
	}

	for i, c := range testCases {
		q := d.createRequest(c.u)
		v := q.Header.Get("Host")

		if v != c.v {
			t.Errorf(`test case %d: expected Host header field value to be "%s", but it is "%s"`, i, c.v, v)
		}
	}
}

func TestDialerCreateRequestHeaders(t *testing.T) {
	d := &Dialer{
		SubProtocols: []string{"chat", "v1"},
	}

	q := d.createRequest(&url.URL{Scheme: "ws", Host: "localhost"})

	v := q.Header.Get("Upgrade")
	e := "websocket"
	if strings.ToLower(v) != e {
		t.Errorf(`expected Upgrade header field value to be "%s", but it is "%s"`, v, e)
	}

	v = q.Header.Get("Connection")
	e = "upgrade"
	if strings.ToLower(v) != e {
		t.Errorf(`expected Connection header field value to be "%s", but it is "%s"`, v, e)
	}

	v = q.Header.Get("Sec-WebSocket-Version")
	e = "13"
	if strings.ToLower(v) != e {
		t.Errorf(`expected Sec-WebSocket-Version header field value to be "%s", but it is "%s"`, v, e)
	}

	v = q.Header.Get("Sec-WebSocket-Protocol")
	e = strings.Join(d.SubProtocols, ", ")
	if strings.ToLower(v) != e {
		t.Errorf(`expected Sec-WebSocket-Protocol header field value to be "%s", but it is "%s"`, v, e)
	}

	l, err := base64.StdEncoding.DecodeString(q.Header.Get("Sec-WebSocket-Key"))

	if err != nil {
		t.Errorf(`unexpected error returned when decoding Sec-WebSocket-Key %s`, err)
	}

	if len(l) != 16 {
		t.Errorf(`expected Sec-WebSocket-Protocol header field value to be '%d' in length, but it is '%d'`, len(l), 16)
	}
}

func TestDialerCreateRequestRequest(t *testing.T) {
	d := &Dialer{}
	u := &url.URL{
		Scheme: "ws",
		Host:   "localhost:8080",
	}

	q := d.createRequest(u)

	if q.URL != u {
		t.Errorf("expected URL instance to be the one provided")
	}

	if q.Method != "GET" {
		t.Errorf(`expected method to be "GET", but it is "%s"`, q.Method)
	}

	if q.Proto != "HTTP/1.1" {
		t.Errorf(`expected http protocol to be "HTTP/1.1", but it is "%s"`, q.Proto)
	}

	if !q.ProtoAtLeast(1, 1) {
		t.Errorf("expected http protocol used to be at least version 1.1, but it is %d.%d", q.ProtoMajor, q.ProtoMinor)
	}

	if q.Host != u.Host {
		t.Errorf(`expected host to be "%s", but it is "%s"`, u.Host, q.Host)
	}
}
