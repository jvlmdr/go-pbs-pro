package grideng

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
	TypeStr string `json:"Type"`
	Header  rawMessage
	Body    rawMessage
}

func makeJSONRequest() jsonRequest {
	return jsonRequest{Header: makeRawMessage(), Body: makeRawMessage()}
}

func (r jsonRequest) Type() string {
	return r.TypeStr
}

func (r jsonRequest) ReadHeader(dst interface{}) error {
	return json.Unmarshal(*r.Header.Message, dst)
}

func (r jsonRequest) ReadBody(dst interface{}) error {
	return json.Unmarshal(*r.Body.Message, dst)
}

// Struct which gets marshalled and unmarshalled for responses.
type jsonResponse struct {
	Header rawMessage
	Body   rawMessage
}

func makeJSONResponse() jsonResponse {
	return jsonResponse{makeRawMessage(), makeRawMessage()}
}

func (r jsonResponse) ReadHeader(dst interface{}) error {
	return json.Unmarshal(*r.Header.Message, dst)
}

func (r jsonResponse) ReadBody(dst interface{}) error {
	return json.Unmarshal(*r.Body.Message, dst)
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
type jsonClientCodec struct{}

func MakeJSONClientCodec() ClientCodec { return jsonClientCodec{} }

func (c jsonClientCodec) WriteRequest(conn io.ReadWriter, typ string, header interface{}, body interface{}) error {
	headerMessage, err := marshal(header)
	if err != nil {
		return err
	}
	bodyMessage, err := marshal(body)
	if err != nil {
		return err
	}
	req := jsonRequest{typ, headerMessage, bodyMessage}

	buf := bufio.NewWriter(conn)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(req); err != nil {
		return err
	}
	return buf.Flush()
}

func (c jsonClientCodec) ReadResponse(conn io.ReadWriter) (Reader, error) {
	var response jsonResponse
	// buf := bufio.NewReader(conn)
	dec := json.NewDecoder(conn)
	if err := dec.Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

//
//
//
type jsonServerCodec struct{}

func MakeJSONServerCodec() ServerCodec { return jsonServerCodec{} }

func (c jsonServerCodec) ReadRequest(conn io.ReadWriter) (RequestReader, error) {
	var request jsonRequest
	// buf := bufio.NewReader(conn)
	dec := json.NewDecoder(conn)
	if err := dec.Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func (c jsonServerCodec) WriteResponse(conn io.ReadWriter, header interface{}, body interface{}) error {
	headerMessage, err := marshal(header)
	if err != nil {
		return err
	}
	bodyMessage, err := marshal(body)
	if err != nil {
		return err
	}
	response := jsonResponse{headerMessage, bodyMessage}

	buf := bufio.NewWriter(conn)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(response); err != nil {
		return err
	}
	return buf.Flush()
}
