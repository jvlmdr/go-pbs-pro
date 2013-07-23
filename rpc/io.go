package main

// There are two types of request: a request to receive input and request to send output.
// The server does not know the type of the incoming request.
type ClientCodec interface {
	// Write request.
	WriteRequest(typ string, header interface{}, body interface{}) error
	// Read response.
	ReadResponse() error
	ReadResponseHeader(interface{}) error
	ReadResponseBody(interface{}) error
}

// The client does know the type of the incoming response.
type ServerCodec interface {
	// Read request.
	ReadRequestType() (string, error)
	ReadRequestHeader(interface{}) error
	ReadRequestBody(interface{}) error
	// Write response.
	WriteResponse(header interface{}, body interface{}) error
}
