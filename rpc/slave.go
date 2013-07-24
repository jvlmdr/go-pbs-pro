package grideng

import (
	"log"
	"net"
	"os"
)

// Never returns.
// If everything succeeds, it calls os.Exit(0).
// Otherwise, it exits with some other status code.
func ExecSlave(task Task, addr, codecName string) {
	codec, err := ClientCodecByName(codecName)
	if err != nil {
		log.Fatalf("Could not create client codec \"%s\": %v", codecName, err)
	}

	// Make request to receive input from server.
	inputResponse := receiveInput(task, addr, codec)
	index := inputResponse.Index
	// Do the thing.
	err = task.Do()
	// Make request to send output to server.
	sendOutput(task, index, err, addr, codec)
	os.Exit(0)
}

// Returns task index.
func receiveInput(task Task, addr string, codec ClientCodec) InputResponseHeader {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln("Could not connect to server:", err)
	}
	defer conn.Close()

	// Send request.
	if err := codec.WriteRequest(conn, InputRequest, nil, nil); err != nil {
		log.Fatalln("Could not write input request:", err)
	}
	// Read response type.
	response, err := codec.ReadResponse(conn)
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

func sendOutput(task Task, index int, taskErr error, addr string, codec ClientCodec) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln("Could not connect to server:", err)
	}
	defer conn.Close()

	// Prepare header.
	header := OutputRequestHeader{Index: index}
	if taskErr != nil {
		header.Error = taskErr.Error()
	}
	// Send request.
	if err := codec.WriteRequest(conn, OutputRequest, header, task.Output()); err != nil {
		log.Fatalln("Could not write request to send output:", err)
	}
	// No data to receive. Done.
}
