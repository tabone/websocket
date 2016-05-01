package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func makeRequestValid(r *http.Request) {
	r.Header.Set("Sec-WebSocket-Version", wsVersion)
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Connection", "upgrade")
	r.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
}

func TestUpgradeResponseWhenInvalidOrigin(t *testing.T) {
	r, err := http.NewRequest("GET", "example.com", nil)

	if err != nil {
		t.Fatal("error occured while creating request:", err)
	}

	w := httptest.NewRecorder()

	h := func(w http.ResponseWriter, r *http.Request) {
		wsr := &Request{
			CheckOrigin: func(r *http.Request) bool {
				return false
			},
		}

		makeRequestValid(r)

		s, err := wsr.Upgrade(w, r)

		if err == nil {
			t.Error("expected Upgrade() to return a OpenError")
		}

		if s != nil {
			t.Error("expected Upgrade() to return a nil Socket instance")
		}
	}

	h(w, r)

	if w.Code != 403 {
		t.Errorf(`expected HTTP Status '403'. '%d' was returned.`, w.Code)
	}
}

func TestUpgradeResponseWhenInvalidWSVersion(t *testing.T) {
	r, err := http.NewRequest("GET", "example.com", nil)

	if err != nil {
		t.Fatal("error occured while creating request:", err)
	}

	w := httptest.NewRecorder()

	h := func(w http.ResponseWriter, r *http.Request) {
		wsr := &Request{}

		makeRequestValid(r)
		r.Header.Set("Sec-WebSocket-Version", "14")

		s, err := wsr.Upgrade(w, r)

		if err == nil {
			t.Error("expected Upgrade() to return a OpenError")
		}

		if s != nil {
			t.Error("expected Upgrade() to return a nil Socket instance")
		}
	}

	h(w, r)

	if w.Code != 426 {
		t.Errorf(`expected HTTP Status '426'. '%d' was returned.`, w.Code)
	}

	if w.Header().Get("Sec-WebSocket-Version") != wsVersion {
		t.Errorf(`expected "Sec-WebSocket-Version" HTTP Header field value to be %s`, wsVersion)
	}
}

func TestUpgradeResponseWhenNotValid(t *testing.T) {
	r, err := http.NewRequest("POST", "example.com", nil)

	if err != nil {
		t.Fatal("error occured while creating request:", err)
	}

	w := httptest.NewRecorder()

	h := func(w http.ResponseWriter, r *http.Request) {
		wsr := &Request{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		makeRequestValid(r)

		s, err := wsr.Upgrade(w, r)

		if err == nil {
			t.Error("expected Upgrade() to return a OpenError.")
		}

		if s != nil {
			t.Error("expected Upgrade() to return a nil Socket instance.")
		}
	}

	h(w, r)

	if w.Code != 400 {
		t.Errorf(`expected HTTP Status '400'. '%d' was returned.`, w.Code)
	}
}

func TestUpgradeGoodRequest(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		wsr := &Request{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		makeRequestValid(r)

		s, err := wsr.Upgrade(w, r)

		if err != nil {
			t.Error("unexpected error from Upgrade():", err)
		}

		if s == nil {
			t.Error("expected Upgrade() to return a non-nil Socket instance")
		}

		if !s.server {
			t.Error("expected socket to have 'server' property set to 'true'")
		}
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	w, err := http.Get(s.URL)

	if err != nil {
		t.Error("unexpected error when requesting the test server:", err)
	}

	if w.StatusCode != 101 {
		t.Errorf("expected HTTP Status to be '101' but it is '%d'", w.StatusCode)
	}

	if w.Header.Get("Upgrade") != "websocket" {
		t.Errorf(`expected "Upgrade" HTTP Header value to be "websocket" but it is "%s"`, w.Header.Get("Upgrade"))
	}

	if w.Header.Get("Connection") != "upgrade" {
		t.Errorf(`expected "Connection" HTTP Header value to be "upgrade" but it is "%s"`, w.Header.Get("Connection"))
	}

	if w.Header.Get("Sec-WebSocket-Accept") != "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" {
		t.Errorf(`expected "Sec-WebSocket-Accept" HTTP Header value to be "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" but it is "%s"`, w.Header.Get("Sec-WebSocket-Accept"))
	}
}

func TestUpgradeWithSubProtocols(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		wsr := &Request{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		makeRequestValid(r)
		wsr.SubProtocol = "one"

		s, err := wsr.Upgrade(w, r)

		if err != nil {
			t.Error("unexpected error from Upgrade():", err)
		}

		if s == nil {
			t.Error("expected Upgrade() to return a non-nil Socket instance")
		}
	}

	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	type testCase struct {
		p string
		v bool
	}

	testCases := []testCase{
		{p: "one, two, three", v: true},
		{p: "two, three", v: false},
		{p: "", v: false},
	}

	for i, c := range testCases {
		_ = i
		r, err := http.NewRequest("GET", s.URL, nil)

		if err != nil {
			t.Error("unexpected error returned while trying to create a request instance:", err)
		}

		if c.p != "" {
			r.Header.Set("Sec-WebSocket-Protocol", c.p)
		}

		l := &http.Client{}
		w, err := l.Do(r)

		if err != nil {
			t.Error("unexpected error returned while trying to create a client instance:", err)
		}

		if c.v {
			v := w.Header.Get("Sec-WebSocket-Protocol")
			if w.Header.Get("Sec-WebSocket-Protocol") != "one" {
				t.Errorf(`expected 'Sec-WebSocket-Protocol' Response Header to be "one", but it is "%v".`, v)
			}
		} else {
			v := w.Header.Get("Sec-WebSocket-Protocol")
			if w.Header.Get("Sec-WebSocket-Protocol") != "" {
				t.Errorf(`expected 'Sec-WebSocket-Protocol' Response Header to be "", but it is "%v".`, v)
			}
		}
	}
}

func TestClientSubProtocols(t *testing.T) {
	r := &http.Request{}

	l := []string{"one", "two", "three"}

	r.Header = make(http.Header)
	r.Header.Set("Sec-WebSocket-Protocol", strings.Join(l, ", "))

	q := &Request{
		request: r,
	}

	p := q.ClientSubProtocols()

	if len(l) != len(p) {
		t.Errorf("The length of the list of header value assigned to Sec-WebSocket-Protocol HTTP Header are not the same. %d != %d", len(l), len(p))
	}

	for _, v := range p {
		k := false
		for _, h := range l {
			if v == h {
				k = true
				break
			}
		}
		if !k {
			t.Errorf(`"%s" was not returned in the slice of Sub Protocols.`, v)
		}
	}
}

func TestClientExtensions(t *testing.T) {
	r := &http.Request{}

	l := []string{"one", "two", "three"}

	r.Header = make(http.Header)
	r.Header.Set("Sec-WebSocket-Extensions", strings.Join(l, ", "))

	q := &Request{
		request: r,
	}

	p := q.ClientExtensions()

	if len(l) != len(p) {
		t.Errorf("The length of the list of header value assigned to Sec-WebSocket-Extensions HTTP Header are not the same. '%d' != '%d'", len(l), len(p))
	}

	for _, v := range p {
		k := false
		for _, h := range l {
			if v == h {
				k = true
				break
			}
		}
		if !k {
			t.Errorf(`"%s" was not returned in the slice of Extensions.`, v)
		}
	}
}
