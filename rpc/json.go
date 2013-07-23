package main

import (
	"bufio"
	"encoding/json"
	"io"
)

type jsonRequest struct {
	Type   string
	Header *json.RawMessage
	Body   *json.RawMessage
}

func makeJSONRequest() jsonRequest {
	return jsonRequest{Header: new(json.RawMessage), Body: new(json.RawMessage)}
}

type jsonResponse struct {
	Header *json.RawMessage
	Body   *json.RawMessage
}

func makeJSONResponse() jsonResponse {
	return jsonResponse{new(json.RawMessage), new(json.RawMessage)}
}

func marshal(x interface{}) (json.RawMessage, error) {
	if x == nil {
		return []byte("null"), nil
	}
	return json.Marshal(x)
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
	rawHeader, err := marshal(header)
	if err != nil {
		return err
	}
	rawBody, err := marshal(body)
	if err != nil {
		return err
	}

	req := jsonRequest{typ, &rawHeader, &rawBody}

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
	return json.Unmarshal(*c.Response.Header, header)
}

func (c jsonClientCodec) ReadResponseBody(body interface{}) error {
	return json.Unmarshal(*c.Response.Body, body)
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
	return json.Unmarshal(*c.Request.Header, header)
}

func (c jsonServerCodec) ReadRequestBody(body interface{}) error {
	return json.Unmarshal(*c.Request.Body, body)
}

func (c jsonServerCodec) WriteResponse(header interface{}, body interface{}) error {
	rawHeader, err := marshal(header)
	if err != nil {
		return err
	}
	rawBody, err := marshal(body)
	if err != nil {
		return err
	}

	response := jsonResponse{&rawHeader, &rawBody}

	buf := bufio.NewWriter(c.Conn)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(response); err != nil {
		return err
	}
	return buf.Flush()
}
