package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSocketReadTextFrame(t *testing.T) {
	payload := "expected payload"

	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.ReadHandler = func(o int, p []byte) {
			if o != OpcodeText {
				t.Errorf("expected opcode to be '%d' but it is '%d'", OpcodeText, o)
			}

			if string(p) != payload {
				t.Errorf(`expected payload to be "%s" but it is "%s"`, p, payload)
			}

			done <- true
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	f := &frame{
		fin:     true,
		opcode:  OpcodeText,
		key:     []byte{1, 1, 1, 1},
		payload: []byte(payload),
	}

	b, err := f.toBytes()

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	c.buf.Write(b)
	if err := c.buf.Flush(); err != nil {
		t.Fatal("unexpected error returned", err)
	}

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketReadBinaryFrame(t *testing.T) {
	payload := "expected payload"

	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.ReadHandler = func(o int, p []byte) {
			if o != OpcodeBinary {
				t.Errorf("expected opcode to be '%d' but it is '%d'", OpcodeBinary, o)
			}

			if string(p) != payload {
				t.Errorf(`expected payload to be "%s" but it is "%s"`, p, payload)
			}

			done <- true
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	f := &frame{
		fin:     true,
		opcode:  OpcodeBinary,
		key:     []byte{1, 1, 1, 1},
		payload: []byte(payload),
	}

	b, err := f.toBytes()

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	c.buf.Write(b)
	if err := c.buf.Flush(); err != nil {
		t.Fatal("unexpected error returned", err)
	}

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketReadPingFrame(t *testing.T) {
	payload := "expected payload"

	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.PingHandler = func(p []byte) {
			if string(p) != payload {
				t.Errorf(`expected payload to be "%s" but it is "%s"`, p, payload)
			}

			done <- true
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	f := &frame{
		fin:     true,
		opcode:  OpcodePing,
		key:     []byte{1, 1, 1, 1},
		payload: []byte(payload),
	}

	b, err := f.toBytes()

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	c.buf.Write(b)
	if err := c.buf.Flush(); err != nil {
		t.Fatal("unexpected error returned", err)
	}

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketReadPongFrame(t *testing.T) {
	payload := "expected payload"

	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.PongHandler = func(p []byte) {
			if string(p) != payload {
				t.Errorf(`expected payload to be "%s" but it is "%s"`, p, payload)
			}

			done <- true
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	f := &frame{
		fin:     true,
		opcode:  OpcodePong,
		key:     []byte{1, 1, 1, 1},
		payload: []byte(payload),
	}

	b, err := f.toBytes()

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	c.buf.Write(b)
	if err := c.buf.Flush(); err != nil {
		t.Fatal("unexpected error returned", err)
	}

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketdefaultPingHandler(t *testing.T) {
	payload := "expected payload"

	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	f := &frame{
		fin:     true,
		opcode:  OpcodePing,
		key:     []byte{1, 1, 1, 1},
		payload: []byte(payload),
	}

	b, err := f.toBytes()

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	c.PongHandler = func(p []byte) {
		if string(p) != payload {
			t.Errorf(`expected payload to be "%s" but it is "%s"`, p, payload)
		}
		done <- true
	}

	go c.Listen()

	c.buf.Write(b)
	if err := c.buf.Flush(); err != nil {
		t.Fatal("unexpected error returned", err)
	}

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketReadInvalidFrame(t *testing.T) {
	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.ReadHandler = func(o int, p []byte) {
			t.Error("unexpected invocation of Read Handler")
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	c.CloseHandler = func(err error) {
		if e, k := err.(*CloseError); k {
			if e.Code != CloseProtocolError {
				t.Errorf("expected Close Error Code to be '%d', but it is '%d'", CloseProtocolError, e.Code)
			}
		} else {
			t.Errorf("expected error instance to be of type *CloseError")
		}
		done <- true
	}

	go c.Listen()

	c.buf.Write([]byte("bad frame"))
	if err := c.buf.Flush(); err != nil {
		t.Fatal("unexpected error returned", err)
	}

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketReadClientUnMaskedFrame(t *testing.T) {
	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.ReadHandler = func(o int, p []byte) {
			t.Errorf("unexpected invocation of Read Handler")
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	f := &frame{
		fin:     true,
		opcode:  OpcodeText,
		payload: []byte("something"),
	}

	b, err := f.toBytes()

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	c.CloseHandler = func(err error) {
		if e, k := err.(*CloseError); k {
			r := "expected payload to be masked"

			if e.Code != CloseProtocolError {
				t.Errorf("expected Close Error Code to be '%d', but it is '%d'", CloseProtocolError, e.Code)
			}

			if e.Reason != r {
				t.Errorf(`expected Close Error Reason to be "%s", but it is "%s"`, r, e.Reason)
			}
		} else {
			t.Errorf("expected error instance to be of type *CloseError")
		}
		done <- true
	}

	go c.Listen()

	c.buf.Write(b)
	if err := c.buf.Flush(); err != nil {
		t.Fatal("unexpected error returned", err)
	}

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketReadServerMaskedFrame(t *testing.T) {
	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		f := &frame{
			fin:     true,
			opcode:  OpcodeText,
			key:     []byte{1, 1, 1, 1},
			payload: []byte("something"),
		}

		b, err := f.toBytes()

		if err != nil {
			t.Fatal("unexpected error returned", err)
		}

		s.CloseHandler = func(err error) {
			if e, k := err.(*CloseError); k {
				r := "expected payload to not be masked"

				if e.Code != CloseProtocolError {
					t.Errorf("expected Close Error Code to be '%d', but it is '%d'", CloseProtocolError, e.Code)
				}

				if e.Reason != r {
					t.Errorf(`expected Close Error Reason to be "%s", but it is "%s"`, r, e.Reason)
				}
			} else {
				t.Errorf("expected error instance to be of type *CloseError")
			}
			done <- true
		}

		s.buf.Write(b)
		if err := s.buf.Flush(); err != nil {
			t.Error("unexpected error returned", err)
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	c.ReadHandler = func(o int, p []byte) {
		t.Errorf("unexpected invocation of Read Handler")
	}

	go c.Listen()

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketClose(t *testing.T) {
	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	c.CloseHandler = func(err error) {
		if e, k := err.(*CloseError); k {
			if e.Code != CloseNormalClosure {
				t.Errorf("expected Close Error Code to be '%d', but it is '%d'", CloseNormalClosure, e.Code)
			}

			if e.Reason != "" {
				t.Errorf(`expected Close Error Reason to be empty, but it is "%s"`, e.Reason)
			}
		} else {
			t.Errorf("expected error instance to be of type *CloseError")
		}
		done <- true
	}

	go c.Listen()

	c.Close()

	select {
	case <-done:
		{
		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketReadEOFError(t *testing.T) {
	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.TCPClose()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	c.CloseHandler = func(err error) {
		if e, k := err.(*CloseError); k {
			r := "abnormal closure"
			if e.Code != CloseAbnormalClosure {
				t.Errorf("expected Close Error Code to be '%d', but it is '%d'", CloseAbnormalClosure, e.Code)
			}

			if e.Reason != r {
				t.Errorf(`expected Close Error Reason to be "%s", but it is "%s"`, r, e.Reason)
			}
		} else {
			t.Errorf("expected error instance to be of type *CloseError")
		}
		done <- true
	}

	go c.Listen()

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketReadTimeoutError(t *testing.T) {
	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 4)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	c.CloseHandler = func(err error) {
		if e, k := err.(*CloseError); k {
			r := "abnormal closure"
			if e.Code != CloseAbnormalClosure {
				t.Errorf("expected Close Error Code to be '%d', but it is '%d'", CloseAbnormalClosure, e.Code)
			}

			if e.Reason != r {
				t.Errorf(`expected Close Error Reason to be "%s", but it is "%s"`, r, e.Reason)
			}
		} else {
			t.Errorf("expected error instance to be of type *CloseError")
		}
		done <- true
	}

	go c.Listen()

	c.SetReadDeadline(time.Now().Add(time.Second * 1))

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketWriteTimeoutErorr(t *testing.T) {
	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 4)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.CloseHandler = func(err error) {
			done <- true
		}

		s.SetWriteDeadline(time.Now().Add(time.Second * 1))

		go s.Listen()

		time.Sleep(time.Second * 2)

		s.WriteMessage(OpcodeText, []byte("something"))
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketWriteFromClient(t *testing.T) {
	payload := "expected payload"

	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		s.ReadHandler = func(o int, p []byte) {
			if o != OpcodeText {
				t.Errorf("expected opcode to be '%d' but it is '%d'", OpcodeText, o)
			}

			if string(p) != payload {
				t.Errorf(`expected payload to be "%s" but it is "%s"`, p, payload)
			}

			done <- true
		}

		s.Listen()
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error was returned", err)
	}

	defer c.TCPClose()

	if err := c.WriteMessage(OpcodeText, []byte(payload)); err != nil {
		t.Fatal("unexpected error returned", err)
	}

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketWriteFromServer(t *testing.T) {
	payload := "expected payload"

	done := make(chan bool)
	timeout := time.NewTicker(time.Second * 2)

	h := func(w http.ResponseWriter, r *http.Request) {
		q := Request{}
		s, err := q.Upgrade(w, r)

		if err != nil {
			t.Fatal("unexpected error was returned", err)
		}

		if err := s.WriteMessage(OpcodeText, []byte(payload)); err != nil {
			t.Fatal("unexpected error was returned", err)
		}
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	d := &Dialer{}
	c, _, err := d.Dial(adaptURL(s.URL))

	if err != nil {
		t.Fatal("unexpected error returned", err)
	}

	defer c.TCPClose()

	c.ReadHandler = func(o int, p []byte) {
		if o != OpcodeText {
			t.Errorf("expected opcode to be '%d' but it is '%d'", OpcodeText, o)
		}

		if string(p) != payload {
			t.Errorf(`expected payload to be "%s" but it is "%s"`, p, payload)
		}

		done <- true
	}

	go c.Listen()

	select {
	case <-done:
		{

		}
	case <-timeout.C:
		{
			t.Error("test case timed out")
		}
	}
}

func TestSocketWriteWhenClosed(t *testing.T) {
	s := &Socket{
		writeMutex: &sync.Mutex{},
	}
	s.state = stateClosed

	if err := s.WriteMessage(1, []byte("test")); err != ErrSocketClosed {
		t.Errorf(`expected error "%s", but got "%v"`, ErrSocketClosed, err)
	}
}

func adaptURL(u string) string {
	return strings.Replace(u, "http://", "ws://", 1)
}
