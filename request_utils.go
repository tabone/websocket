package websocket

import (
	"encoding/base64"
	"net/http"
	"strings"
)

/*
	validateRequest is used to determine whether the client handshake request
	conforms with the WebSocket spec. When it doesn't the server should respond
	with an HTTP Status 400 Bad Request.

	Note that this method doesn't validate the websocket version
	("Sec-WebSocket-Version" HTTP Header Field)  and origin ("Origin" HTTP
	Header Field) since these require specific HTTP Status Code (427 and 403
	respectively).

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
	          https://tools.ietf.org/html/rfc6455#section-4.2.2
*/
func validateRequest(r *http.Request) *OpenError {
	validations := []func(*http.Request) *OpenError{
		// Check HTTP version to be at least v1.1.
		validateRequestVersion,
		// Check HTTP method to be 'GET'.
		validateRequestMethod,
		// Validate 'Upgrade' header field.
		validateRequestUpgradeHeader,
		// Validate 'Connection' header field.
		validateRequestConnectionHeader,
		// Validate 'Sec-WebSocket-Key' header field.
		validateRequestSecWebsocketKeyHeader,
	}

	for _, v := range validations {
		if err := v(r); err != nil {
			return err
		}
	}

	return nil
}

/*
	validateRequestVersion verifies that the HTTP Version used in the client's
	opening handshake request is at least v1.1.

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
*/
func validateRequestVersion(r *http.Request) *OpenError {
	if !r.ProtoAtLeast(1, 1) {
		return &OpenError{Reason: `HTTP must be v1.1 or higher`}
	}
	return nil
}

/*
	validateRequestMethod verifies that the HTTP Method used in the client's
	opening handshake request is 'GET'.

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
*/
func validateRequestMethod(r *http.Request) *OpenError {
	if r.Method != "GET" {
		return &OpenError{Reason: `HTTP method must be "GET"`}
	}
	return nil
}

/*
	validateRequestUpgradeHeader verifies that the Upgrade HTTP Header value in the
	client's opening handshake request is "websocket".

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
*/
func validateRequestUpgradeHeader(r *http.Request) *OpenError {
	h := r.Header.Get("Upgrade")

	if strings.ToLower(h) != "websocket" {
		return &OpenError{Reason: `"Upgrade" Header should have the value of "websocket"`}
	}

	return nil
}

/*
	validateRequestConnectionHeader verfies that the Connection HTTP Header value in
	the client's opening handshake request is "upgrade".

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
*/
func validateRequestConnectionHeader(r *http.Request) *OpenError {
	h := r.Header.Get("Connection")

	if strings.ToLower(h) != "upgrade" {
		return &OpenError{Reason: `"Connection" Header should have the value of "upgrade"`}
	}

	return nil
}

/*
	validateRequestSecWebsocketKeyHeader verifies that the Sec-WebSocket-Key HTTP Header value in
	the client's opening handshake request is of length 16 when base64 decoded.

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
*/
func validateRequestSecWebsocketKeyHeader(r *http.Request) *OpenError {
	h := r.Header.Get("Sec-WebSocket-Key")
	d, err := base64.StdEncoding.DecodeString(h)

	// Check for decoding errors.
	if err != nil {
		return &OpenError{Reason: `an error had occured while validating "Sec-WebSocket-Key" header`}
	}

	// Check that the length of the decoded Sec-WebSocket-Key value is 16
	// (bytes).
	if len(d) != 16 {
		return &OpenError{Reason: `"Sec-WebSocket-Key" must be 16 bytes in length when decoded`}
	}

	return nil
}

/*
	validateWSVersionHeader verifies that the Sec-WebSocket-Verion HTTP Header
	value in the client's opening handshake request is "13".

	Ref Spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
*/
func validateWSVersionHeader(r *http.Request) *OpenError {
	if r.Header.Get("Sec-WebSocket-Version") != wsVersion {
		return &OpenError{Reason: "upgrade required"}
	}

	return nil
}

/*
	checkOrigin is the default CheckOrigin handler used by the Request struct.
	This method will allow requests that are either coming from a non-browser
	client (Origin HTTP Header field omitted) or are not cross origin requests.

	Ref spec: https://tools.ietf.org/html/rfc6455#section-4.2.1
*/
func checkOrigin(r *http.Request) bool {
	h := r.Header.Get("Origin")

	if strings.HasPrefix(h, "http://") {
		h = strings.Replace(h, "http://", "", 1)
	} else if strings.HasPrefix(h, "https://") {
		h = strings.Replace(h, "https://", "", 1)
	}

	return h == "" || h == r.Host
}
