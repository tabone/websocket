package websocket

import (
	"net/http"
	"testing"
)

func TestValidateRequestVersion(t *testing.T) {
	r := &http.Request{}
	r.Header = make(http.Header)

	type testCase struct {
		a int
		i int
		r bool
	}

	testCases := []testCase{
		// HTTP v1.1 should be valid.
		{a: 1, i: 1, r: true},
		// HTTP v2.1 should be valid.
		{a: 2, i: 1, r: true},
		// HTTP v1.0 should be not valid.
		{a: 1, i: 0, r: false},
		// HTTP v0.1 should be not valid.
		{a: 0, i: 1, r: false},
	}

	for i, c := range testCases {
		r.ProtoMajor = c.a
		r.ProtoMinor = c.i

		err := validateRequestVersion(r)

		if c.r && err != nil {
			t.Errorf(`test case %d: unexpected error retured for "v%d.%d"`, i, c.a, c.i)
		}

		if !c.r && err == nil {
			t.Errorf(`test case %d: expected an error for "v%d.%d"`, i, c.a, c.i)
		}
	}
}

func TestValidateRequestMethod(t *testing.T) {
	r := &http.Request{}
	r.Header = make(http.Header)

	type testCase struct {
		m string
		r bool
	}

	testCases := []testCase{
		// HTTP GET should be valid.
		{m: "GET", r: true},
		// HTTP POST should be not valid.
		{m: "POST", r: false},
	}

	for i, c := range testCases {
		r.Method = c.m

		err := validateRequestMethod(r)

		if c.r && err != nil {
			t.Errorf(`test case %d: unexpected error retured for "%s" request`, i, c.m)
		}

		if !c.r && err == nil {
			t.Errorf(`test case %d: expected an error for  "%s" request`, i, c.m)
		}
	}
}

func TestValidateRequestUpgradeHeader(t *testing.T) {
	r := &http.Request{}
	r.Header = make(http.Header)

	type testCase struct {
		v string
		r bool
	}

	testCases := []testCase{
		// When value is "websocket" should be valid.
		{v: "websocket", r: true},
		// When value is "webSocket" should be valid.
		{v: "webSocket", r: true},
		// When value is not "websocket" should not be valid.
		{v: "ValueOtherThanWebsocket", r: false},
	}

	for i, c := range testCases {
		r.Header.Set("Upgrade", c.v)

		err := validateRequestUpgradeHeader(r)

		if c.r && err != nil {
			t.Errorf(`test case %d: unexpected error retured for "%s"`, i, c.v)
		}

		if !c.r && err == nil {
			t.Errorf(`test case %d: expected an error for "%s"`, i, c.v)
		}
	}
}

func TestValidateRequestConnectionHeader(t *testing.T) {
	r := &http.Request{}
	r.Header = make(http.Header)

	type testCase struct {
		v string
		r bool
	}

	testCases := []testCase{
		// When value is "upgrade" should be valid.
		{v: "upgrade", r: true},
		// When value is "Upgrade" should be valid.
		{v: "Upgrade", r: true},
		// When value is not "upgrade" should not be valid.
		{v: "ValueOtherThanUpgrade", r: false},
	}

	for i, c := range testCases {
		r.Header.Set("Connection", c.v)

		err := validateRequestConnectionHeader(r)

		if c.r && err != nil {
			t.Errorf(`test case %d: unexpected error retured for "%s"`, i, c.v)
		}

		if !c.r && err == nil {
			t.Errorf(`test case %d: expected an error for "%s"`, i, c.v)
		}
	}
}

func TestValidateRequestSecWebsocketKeyHeader(t *testing.T) {
	r := &http.Request{}
	r.Header = make(http.Header)

	type testCase struct {
		v string
		r bool
	}

	testCases := []testCase{
		// Valid key.
		{v: "FlBPpXKmN36AUZxV0tYHYw==", r: true},
		// Invalid decoded length.
		{v: "InvalidKey==", r: false},
		// Invalid encoded data.
		{v: "InvalidKeyError", r: false},
	}

	for i, c := range testCases {
		r.Header.Set("Sec-WebSocket-Key", c.v)

		err := validateRequestSecWebsocketKeyHeader(r)

		if c.r && err != nil {
			t.Errorf(`test case %d: unexpected error retured for "%s"`, i, c.v)
		}

		if !c.r && err == nil {
			t.Errorf(`test case %d: expected an error for "%s"`, i, c.v)
		}
	}
}

func TestValidateWSVersionHeader(t *testing.T) {
	r := &http.Request{}
	r.Header = make(http.Header)

	type testCase struct {
		v string
		r bool
	}

	testCases := []testCase{
		// Valid when value is the same as the version of the ws supported.
		{v: wsVersion, r: true},
		// Not valid when value is not the same as the version of the ws
		// supported.
		{v: "14", r: false},
	}

	for i, c := range testCases {
		r.Header.Set("Sec-WebSocket-Version", c.v)

		err := validateWSVersionHeader(r)

		if c.r && err != nil {
			t.Errorf(`test case %d: unexpected error retured for "%s"`, i, c.v)
		}

		if !c.r && err == nil {
			t.Errorf(`test case %d: expected an error for "%s"`, i, c.v)
		}
	}
}

func TestCheckOrigin(t *testing.T) {
	r := &http.Request{}
	r.Header = make(http.Header)
	r.Host = "example.com:8080"

	type testCase struct {
		v string
		r bool
	}

	testCases := []testCase{
		// Valid when origin is omitted (non-browser client).
		{v: "", r: true},
		// Valid when same origin.
		{v: r.Host, r: true},
		{v: "example.com:8080", r: true},
		{v: "http://example.com:8080", r: true},
		{v: "https://example.com:8080", r: true},
	}

	for i, c := range testCases {
		r.Header.Set("Origin", c.v)

		if checkOrigin(r) != c.r {
			t.Errorf(`Test Case %d: Expected checkOrigin() to return '%t' when 'Origin' header == "%s" and Host is at "%s".`, i, c.r, c.v, r.Host)
		}
	}
}
