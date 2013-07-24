package rpc

import (
	"fmt"
	"log"
	"net"
	"os"
)

// Never returns.
// If everything succeeds, it calls os.Exit(0).
// Otherwise, it exits with some other status code.
func ExecSlave(task Task, addr, port, codec string) {
	// Make request to receive input from server.
	inputResponse := receiveInput(task, addr, port, codec)
	index := inputResponse.Index
	// Do the thing.
	err := task.Do()
	// Make request to send output to server.
	sendOutput(task, index, err, addr, port, codec)
	os.Exit(0)
}

// Returns task index.
func receiveInput(task Task, addr, port, codecName string) InputResponseHeader {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		log.Fatalln("Could not connect to server:", err)
	}
	defer conn.Close()

	// Get codec.
	codec, err := MakeClientCodecByName(codecName, conn)
	if err != nil {
		log.Fatalf("Could not create client codec \"%s\": %v", codecName, err)
	}
	// Send request.
	if err := codec.WriteRequest(InputRequest, nil, nil); err != nil {
		log.Fatalln("Could not write input request:", err)
	}
	// Read response type.
	response, err := codec.ReadResponse()
	if err != nil {
		log.Fatalln("Could not read response to input request:", err)
	}
	// Read response header.
	var header InputResponseHeader
	if err := response.ReadHeader(&header); err != nil {
		log.Fatalln("Could not read header of response to input request:", err)
	}
	log.Println("Task index:", header.Index)
	// Read response body.
	if err := response.ReadBody(task.Input()); err != nil {
		log.Fatalln("Could not read body of response to input request:", err)
	}
	return header
}

func sendOutput(task Task, index int, taskErr error, addr, port, codecName string) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		log.Fatalln("Could not connect to server:", err)
	}
	defer conn.Close()

	// Get codec.
	codec, err := MakeClientCodecByName(codecName, conn)
	if err != nil {
		log.Fatalf("Could not create client codec \"%s\": %v", codecName, err)
	}
	// Prepare header.
	header := OutputRequestHeader{Index: index}
	if taskErr != nil {
		header.Error = taskErr.Error()
	}
	// Send request.
	if err := codec.WriteRequest(OutputRequest, header, task.Output()); err != nil {
		log.Fatalln("Could not write request to send output:", err)
	}
	// No data to receive. Done.
}
