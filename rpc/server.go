package rpc

import (
	"errors"
	"fmt"
	"io"
	"log"
)

func handle(m Map, conn io.ReadWriter, codecName string, todo <-chan int, errs chan<- error) {
	// Get codec.
	codec, err := MakeServerCodecByName(codecName, conn)
	if err != nil {
		log.Printf("Could not create server codec \"%s\": %v", codecName, err)
		return
	}

	// Read request type.
	request, err := codec.ReadRequest()
	if err != nil {
		log.Println("Could not read request:", err)
		return
	}

	switch t := request.Type(); t {
	default:
		log.Printf(`Unknown request type "%s"`, t)
	case InputRequest:
		// No more information to read since input request has no content.
		// Get index of new job from channel for thread safety.
		index := <-todo
		// Prepare header and send input.
		header := InputResponseHeader{index}
		if err := codec.WriteResponse(header, m.Input(index)); err != nil {
			log.Println("Could not write response to input request:", err)
			return
		}
	case OutputRequest:
		// Read header and body of request to send output.
		var header OutputRequestHeader
		if err := request.ReadHeader(&header); err != nil {
			fmt.Println("Could not read header of request to send output:", err)
			return
		}
		index := header.Index
		// Read body into map output.
		// Is accessing m.Output thread-safe?
		// Perhaps the output should be passed down the channel?
		// But the type is unknown...
		if err := request.ReadBody(m.Output(index)); err != nil {
			fmt.Println("Could not read body of request to send output:", err)
			return
		}
		// No need to send response to client.
		// Report that a job is finished.
		var err error = nil
		if len(header.Error) != 0 {
			err = errors.New(header.Error)
		}
		errs <- err
	}
}
