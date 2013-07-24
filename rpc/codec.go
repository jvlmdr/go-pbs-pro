package grideng

import "io"

const (
	InputRequest  = "input"
	OutputRequest = "output"
)

// There are two types of request: a request to receive input and request to send output.
// The server does not know the type of the incoming request.
// Codecs have no state.
type ClientCodec interface {
	// Write request.
	WriteRequest(conn io.ReadWriter, typ string, header interface{}, body interface{}) error
	// Read response.
	ReadResponse(conn io.ReadWriter) (Reader, error)
}

type Reader interface {
	ReadHeader(interface{}) error
	ReadBody(interface{}) error
}

type RequestReader interface {
	Type() string
	Reader
}

// The client does know the type of the incoming response.
// Codecs have no state.
type ServerCodec interface {
	// Read request.
	ReadRequest(conn io.ReadWriter) (RequestReader, error)
	// Write response.
	WriteResponse(conn io.ReadWriter, header interface{}, body interface{}) error
}

// A request to send output contains an index, and possibly an error (plus the body).
type OutputRequestHeader struct {
	Index int
	Error string
}

// A response to send input contains the index (plus the body).
type InputResponseHeader struct {
	Index int
}
