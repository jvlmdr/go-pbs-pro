package main

// Ripped off from net/rpc.

//	type ServerCodec struct {
//		ReadRequestHeader(*Request) error
//		ReadRequestBody(interface{}) error
//		WriteResponse(*Response, interface{}) error
//		Close() error
//	}
//
//	type ClientCodec struct {
//		WriteRequest(*Request, interface{}) error
//		ReadResponseHeader(*Response) error
//		ReadResponseBody(interface{}) error
//		Close() error
//	}
//
//	type RequestType int
//
//	const (
//		InputRequest RequestType = iota
//		OutputRequest
//	)
//
//	type RequestHeader struct {
//		Type RequestType
//	}
//
//	type jsonRequest struct {
//		Type RequestType
//		Data *json.RawMessage
//	}
//
//
//
