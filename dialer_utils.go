package websocket

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// validateResponse is used to determine whether the servers handshake request
// conforms with the WebSocket spec. When it doesn't the client fails the
// websocket connection.
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.1
func validateResponse(r *http.Response) *OpenError {
	validations := []func(*http.Response) *OpenError{
		validateResponseStatus,
		validateResponseUpgradeHeader,
		validateResponseConnectionHeader,
		validateResponseSecWebsocketAcceptHeader,
	}

	for _, v := range validations {
		if err := v(r); err != nil {
			return err
		}
	}

	return nil
}

// validateResponseStatus verifies that status code of the server's opening
// handshake response is '101'. If it is not, it means that the handshake has
// been rejected and thus the endpoints are still communicating using http.
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.1
func validateResponseStatus(r *http.Response) *OpenError {
	if r.StatusCode != 101 {
		return &OpenError{
			Reason: "http status not 101",
		}
	}
	return nil
}

// validateResponseUpgradeHeader verifies that the Upgrade HTTP Header value
// in the servers's opening handshake response is "websocket".
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.1
func validateResponseUpgradeHeader(r *http.Response) *OpenError {
	if strings.ToLower(r.Header.Get("Upgrade")) != "websocket" {
		return &OpenError{
			Reason: `"Upgrade" Header should have the value of "websocket"`,
		}
	}
	return nil
}

// validateResponseConnectionHeader verifies that the Connection HTTP Header
// value in the servers's opening handshake response is "upgrade".
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.1
func validateResponseConnectionHeader(r *http.Response) *OpenError {
	if strings.ToLower(r.Header.Get("Connection")) != "upgrade" {
		return &OpenError{
			Reason: `"Connection" Header should have the value of "upgrade"`,
		}
	}
	return nil
}

// validateResponseSecWebsocketAcceptHeader verifies that the
// Sec-WebSocket-Accept HTTP Header value in the server's opening handshake
// response is the base64-encoded SHA-1 of the concatenation of the
// Sec-WebSocket-Key value (sent with the opening handshake request) (as a
// string, not base64-decoded) with the websocket accept key.
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.1
func validateResponseSecWebsocketAcceptHeader(r *http.Response) *OpenError {
	if r.Header.Get("Sec-WebSocket-Accept") != makeAcceptKey(r.Request.Header.Get("Sec-Websocket-Key")) {
		return &OpenError{
			Reason: `challenge key failure`,
		}
	}
	return nil
}

// validateResponseSecWebsocketProtocol verifies that the sub protocol the
// server has agreed to use (Sec-WebSocket-Protocol Header) was in the list the
// client has sent in the opening handshake request.
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.1
func validateResponseSecWebsocketProtocol(r *http.Response) *OpenError {
	// Sub protocols sent by the client.
	c := headerToSlice(r.Request.Header.Get("Sec-WebSocket-Protocol"))
	// Sub protocol the server has agreed to use.
	s := r.Header.Get("Sec-WebSocket-Protocol")

	// If the server hasn't agreed to use anything, stop process.
	if len(s) == 0 {
		return nil
	}

	// Loop through the lists of sub protocols the client has sent in its
	// opening handshake request and if the sub protocol the server argeed to
	// use is found stop the process.
	for _, cv := range c {
		if cv == s {
			return nil
		}
	}

	// At this point the server has agreed to use a sub protocol which the
	// client doesn't support and thus return an error.
	return &OpenError{
		Reason: `server choose a sub protocol which was not in the list sent by the client`,
	}
}

// makeChallengeKey is used to generate the key to be sent with the client's
// opening handshake using the Sec-Websocket-Key header field.
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.1
func makeChallengeKey() string {
	// return Base64 encode version of the byte generated.
	return base64.StdEncoding.EncodeToString(randomByteSlice(4))
}

// parseURL is used to parse the URL string provided and verifies that it
// conforms with the websocket spec. If it does it will create and return a URL
// instance representing the URL string provided.
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-3
func parseURL(u string) (*url.URL, error) {
	// Parse scheme.
	if err := parseURLScheme(&u); err != nil {
		return nil, err
	}

	// Create URL Instance.
	l, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	// Parse Host.
	if err := parseURLHost(l); err != nil {
		return nil, err
	}

	return l, nil
}

// parseURLScheme is used to parse the Scheme portion of a URL string. If the
// scheme provided is not a valid websocket scheme an error is returned. If no
// scheme is given it will be defaulted to "ws".
// 
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-3
func parseURLScheme(u *string) error {
	// Regex to retrieve Scheme portion of a URL string.
	re := regexp.MustCompile("^([a-zA-Z]+)://")
	m := re.FindStringSubmatch(*u)

	// If m is smaller than 2 it means that the user hasn't provided one and
	// thus the default sheme (ws) is used.
	if len(m) < 2 {
		*u = "ws://" + *u
		return nil
	}

	// If a sheme was captured, make sure it is valid.
	if !schemeValid(m[1]) {
		return errors.New("invalid scheme: " + m[1])
	}

	return nil
}

// parseURLHost is used to parse the Host portion of a URL instance to
// determine whether it has a port or not. When no port is found this method
// will assign a port based on the URL instance scheme (ws = 22, wss = 443). If
// the scheme is not a valid scheme for websocket an error is returned.
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-3
func parseURLHost(u *url.URL) error {
	// If scheme is invalid throw an error
	if !schemeValid(u.Scheme) {
		return errors.New("invalid scheme: " + u.Scheme)
	}

	// Regex to retrieve the Port portion of the URL.
	re := regexp.MustCompile(":(\\d+)")
	m := re.FindStringSubmatch(u.Host)

	// If the length of m is greater than or equals to 2 it means that there is
	// a submatch, meaning that the user has provided a port and thus there is
	// no need to include the default ports.
	if len(m) >= 2 {
		return nil
	}

	// Based on the scheme, set the port.
	switch u.Scheme {
	case "ws":
		{
			u.Host += ":22"
		}
	case "wss":
		{
			u.Host += ":443"
		}
	}

	return nil
}

// schemeValid is used to determine whether the scheme provided is a valid
// scheme for the websocket protocol.
func schemeValid(s string) bool {
	return s == "ws" || s == "wss"
}
