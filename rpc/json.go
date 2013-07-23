package main

import (
	"bufio"
	"encoding/json"
	"io"
)

// Wrap json.RawMessage in struct so that even if the message is a nil pointer, the JSON string will be {} not empty.
type rawMessage struct {
	Message *json.RawMessage
}

func makeRawMessage() rawMessage { return rawMessage{new(json.RawMessage)} }

// Struct which gets marshalled and unmarshalled for requests.
type jsonRequest struct {
	Type   string
	Header rawMessage
	Body   rawMessage
}

func makeJSONRequest() jsonRequest {
	return jsonRequest{Header: makeRawMessage(), Body: makeRawMessage()}
}

// Struct which gets marshalled and unmarshalled for responses.
type jsonResponse struct {
	Header rawMessage
	Body   rawMessage
}

func makeJSONResponse() jsonResponse {
	return jsonResponse{makeRawMessage(), makeRawMessage()}
}

// Calls json.Marshal but returns a nil pointer if x is nil.
func marshal(x interface{}) (rawMessage, error) {
	if x == nil {
		return rawMessage{}, nil
	}
	var message json.RawMessage
	message, err := json.Marshal(x)
	if err != nil {
		return rawMessage{}, err
	}
	return rawMessage{&message}, err
}

//
//
//
type jsonClientCodec struct {
	Conn io.ReadWriter
	// Store response as it is being read.
	Response jsonResponse
}

func NewJSONClientCodec(conn io.ReadWriter) ClientCodec {
	return &jsonClientCodec{conn, makeJSONResponse()}
}

func (c jsonClientCodec) WriteRequest(typ string, header interface{}, body interface{}) error {
	headerMessage, err := marshal(header)
	if err != nil {
		return err
	}
	bodyMessage, err := marshal(body)
	if err != nil {
		return err
	}
	req := jsonRequest{typ, headerMessage, bodyMessage}

	buf := bufio.NewWriter(c.Conn)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(req); err != nil {
		return err
	}
	return buf.Flush()
}

func (c *jsonClientCodec) ReadResponse() error {
	// buf := bufio.NewReader(c.Conn)
	dec := json.NewDecoder(c.Conn)
	if err := dec.Decode(&c.Response); err != nil {
		return err
	}
	return nil
}

func (c jsonClientCodec) ReadResponseHeader(header interface{}) error {
	return json.Unmarshal(*c.Response.Header.Message, header)
}

func (c jsonClientCodec) ReadResponseBody(body interface{}) error {
	return json.Unmarshal(*c.Response.Body.Message, body)
}

//
//
//
type jsonServerCodec struct {
	Conn io.ReadWriteCloser
	// Store request as it is being read.
	Request jsonRequest
}

func NewJSONServerCodec(conn io.ReadWriteCloser) ServerCodec {
	return &jsonServerCodec{Conn: conn}
}

func (c *jsonServerCodec) ReadRequestType() (string, error) {
	// buf := bufio.NewReader(c.Conn)
	dec := json.NewDecoder(c.Conn)
	if err := dec.Decode(&c.Request); err != nil {
		return string(0), err
	}
	return c.Request.Type, nil
}

func (c jsonServerCodec) ReadRequestHeader(header interface{}) error {
	return json.Unmarshal(*c.Request.Header.Message, header)
}

func (c jsonServerCodec) ReadRequestBody(body interface{}) error {
	return json.Unmarshal(*c.Request.Body.Message, body)
}

func (c jsonServerCodec) WriteResponse(header interface{}, body interface{}) error {
	headerMessage, err := marshal(header)
	if err != nil {
		return err
	}
	bodyMessage, err := marshal(body)
	if err != nil {
		return err
	}
	response := jsonResponse{headerMessage, bodyMessage}

	buf := bufio.NewWriter(c.Conn)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(response); err != nil {
		return err
	}
	return buf.Flush()
}
