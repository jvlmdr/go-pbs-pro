package rpc

import (
	"errors"
	"io"
)

var (
	ErrCodecName = errors.New("unknown codec")
)

func MakeClientCodecByName(name string, conn io.ReadWriter) (ClientCodec, error) {
	switch name {
	case "json":
		codec := MakeJSONClientCodec(conn)
		return codec, nil
	default:
	}
	return nil, ErrCodecName
}

func MakeServerCodecByName(name string, conn io.ReadWriter) (ServerCodec, error) {
	switch name {
	case "json":
		codec := MakeJSONServerCodec(conn)
		return codec, nil
	default:
	}
	return nil, ErrCodecName
}
