package grideng

import "errors"

const (
	JSONCodec = "json"
	GobCodec  = "gob"
)

var (
	ErrCodecName = errors.New("unknown codec")
)

func ClientCodecByName(name string) (ClientCodec, error) {
	switch name {
	case JSONCodec:
		return jsonClientCodec{}, nil
	default:
	}
	return nil, ErrCodecName
}

func ServerCodecByName(name string) (ServerCodec, error) {
	switch name {
	case JSONCodec:
		return jsonServerCodec{}, nil
	default:
	}
	return nil, ErrCodecName
}
