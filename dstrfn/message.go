package dstrfn

import (
	"encoding/json"
)

const (
	recvType = "recv"
	sendType = "send"
)

// Returns a client request to receive input.
func inputReq() *request {
	return &request{recvType, json.RawMessage("null")}
}

// Describes a server response to send input.
type inputResp struct {
	Index int
	X     interface{}
	P     interface{}
}

// Describes a client request to send output.
// Use a string pointer for its ability to represent the nil error.
// Cannot use error because it is an interface and
// therefore not compatible with Marshal/Unmarshal.
type outputReq struct {
	Index int
	Y     interface{}
	Err   *string
}

// Returns a generic request.
func (r *outputReq) Generic() *request {
	body, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return &request{sendType, json.RawMessage(body)}
}

type request struct {
	Type string
	Body json.RawMessage
}
