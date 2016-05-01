package websocket

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"testing"
)

func TestParseURLScheme(t *testing.T) {
	type testCase struct {
		u string
		f string
	}

	testCases := []testCase{
		{u: "ws://localhost:8080", f: "ws://localhost:8080"},
		{u: "wss://localhost:8080", f: "wss://localhost:8080"},
		{u: "localhost:8080", f: "ws://localhost:8080"},
	}

	for i, c := range testCases {
		n := c.u

		if err := parseURLScheme(&n); err != nil {
			t.Errorf("test case %d: unexpected error was returned %s", i, err)
		}

		if n != c.f {
			t.Errorf(`test case %d: expected url to be "%s", but it is "%s"`, i, c.f, n)
		}
	}
}

func TestParseURLSchemeError(t *testing.T) {
	u := "http://localhost:8080"

	if err := parseURLScheme(&u); err == nil {
		t.Error("expected an error for", u)
	}
}

func TestParseURLHost(t *testing.T) {
	type testCase struct {
		u *url.URL
		h string
	}

	testCases := []testCase{
		{u: &url.URL{Scheme: "ws", Host: "localhost:80"}, h: "localhost:80"},
		{u: &url.URL{Scheme: "wss", Host: "localhost:80"}, h: "localhost:80"},
		{u: &url.URL{Scheme: "ws", Host: "localhost"}, h: "localhost:22"},
		{u: &url.URL{Scheme: "wss", Host: "localhost"}, h: "localhost:443"},
	}

	for i, c := range testCases {
		if err := parseURLHost(c.u); err != nil {
			t.Errorf("test case %d: unexpected error was returned %s", i, err)
		}

		if c.u.Host != c.h {
			t.Errorf(`test case %d: expected host to be "%s", but it is "%s"`, i, c.h, c.u.Host)
		}
	}
}

func TestParseURLHostError(t *testing.T) {
	u := &url.URL{
		Scheme: "http",
		Host:   "localhost",
	}

	if err := parseURLHost(u); err == nil {
		t.Errorf("expected an error to be returned")
	}
}

func TestMakeChallengeKey(t *testing.T) {
	k := makeChallengeKey()
	b, err := base64.StdEncoding.DecodeString(k)

	if err != nil {
		t.Errorf("unexpected error was returned while decoding value: %s", err)
	}

	if len(b) != 16 {
		t.Errorf("expected length of decoded challenge key to be 16, but it is %d", len(b))
	}
}

func TestValidateResponseStatus(t *testing.T) {
	type testCase struct {
		s int
		e bool
	}

	testCases := []testCase{
		{s: 101, e: false},
		{s: 200, e: true},
	}

	for i, c := range testCases {
		r := &http.Response{
			StatusCode: c.s,
		}

		err := validateResponseStatus(r)

		if c.e && err == nil {
			t.Errorf(`test case %d: expected an error for '%d'`, i, c.s)
		}

		if !c.e && err != nil {
			t.Errorf(`test case %d: unexpected error returned for '%d'`, i, c.s)
		}
	}
}

func TestValidateResponseUpgradeHeader(t *testing.T) {
	type testCase struct {
		v string
		e bool
	}

	testCases := []testCase{
		{v: "websocket", e: false},
		{v: "WebSocket", e: false},
		{v: "wrong", e: true},
	}

	for i, c := range testCases {
		r := &http.Response{
			Header: make(http.Header),
		}

		r.Header.Add("Upgrade", c.v)
		err := validateResponseUpgradeHeader(r)

		if c.e && err == nil {
			t.Errorf(`test case %d: expected an error for "%s"`, i, c.v)
		}

		if !c.e && err != nil {
			t.Errorf(`test case %d: unexpected error returned for "%s"`, i, c.v)
		}
	}
}

func TestValidateResponseConnectionHeader(t *testing.T) {
	type testCase struct {
		v string
		e bool
	}

	testCases := []testCase{
		{v: "upgrade", e: false},
		{v: "UpgrADE", e: false},
		{v: "wrong", e: true},
	}

	for i, c := range testCases {
		r := &http.Response{
			Header: make(http.Header),
		}

		r.Header.Add("Connection", c.v)
		err := validateResponseConnectionHeader(r)

		if c.e && err == nil {
			t.Errorf(`test case %d: expected an error for "%s"`, i, c.v)
		}

		if !c.e && err != nil {
			t.Errorf(`test case %d: unexpected error returned for "%s"`, i, c.v)
		}
	}
}

func TestValidateResponseSecWebsocketProtocol(t *testing.T) {
	type testCase struct {
		c string
		s string
		e bool
	}

	testCases := []testCase{
		{c: "client, v1", s: "", e: false},
		{c: "client, v1", s: "v1", e: false},
		{c: "client, v1", s: "v2", e: true},
	}

	for i, c := range testCases {
		// Headers sent by client
		hq := make(http.Header)
		hq.Set("Sec-WebSocket-Protocol", c.c)

		// Headers sent by server
		hr := make(http.Header)
		hr.Set("Sec-WebSocket-Protocol", c.s)

		q := &http.Request{
			Header: hq,
		}

		r := &http.Response{
			Header:  hr,
			Request: q,
		}

		err := validateResponseSecWebsocketProtocol(r)

		if c.e && err == nil {
			t.Errorf(`test case %d: expected an error when the client sent "%s" as supported protocols and the server agreed to use "%s"`, i, c.c, c.s)
		}

		if !c.e && err != nil {
			t.Errorf(`test case %d: unexpected error was returned when the client sent "%s" as supported protocols and the server agreed to use "%s"`, i, c.c, c.s)
		}
	}
}
