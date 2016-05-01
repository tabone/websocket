package websocket

// mask is used to mask or unmask an array of bytes. It accepts two arguments,
// p the data that will be masked (usually the application data), k the masking
// key.
// 
// From spec: https://tools.ietf.org/html/rfc6455#section-5.3
func mask(p, k []byte) {
	for i := range p {
		p[i] ^= k[i%4]
	}
}

// opcodeExist returns whether the opcode number provided as an argument is a
// valid opcode or not.
func opcodeExist(i int) bool {
	switch i {
	case OpcodeContinuation, OpcodeText, OpcodeBinary, OpcodeClose, OpcodePing, OpcodePong:
		{
			return true
		}
	}
	return false
}

// validateKey returns whether the masking key is a valid key or not. Note that
// a masking key can either be of length 0 or 4.
//
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.2
func validateKey(k []byte) bool {
	return len(k) == 0 || len(k) == 4
}

// validatePayload returns whether the payload data is valid or not. Note that
// the maximum size of payload data can be 9223372036854775807 bits.
// 
// Ref Spec: https://tools.ietf.org/html/rfc6455#section-5.2
func validatePayload(p []byte) bool {
	return len(p) <= 9223372036854775807
}
