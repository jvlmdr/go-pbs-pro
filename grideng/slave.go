package grideng

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
)

// If the process is a slave, this function never returns.
func ExecIfSlave() {
	if len(slaveTask) == 0 {
		return
	}

	// Look up task by name.
	sub, there := tasks[slaveTask]
	if !there {
		panic(fmt.Errorf("task not found: %#v", slaveTask))
	}

	slave(sub.Task)
	os.Exit(0)
}

func slave(task Task) {
	// Request input from the master.
	x := task.NewInput()
	p := task.NewConfig()
	index, err := receiveInput(addrStr, x, p)
	if err != nil {
		panic(err)
	}

	y, taskerr := task.Func(x, p)
	if err := sendOutput(addrStr, index, y, taskerr); err != nil {
		panic(err)
	}
}

// Populates the values referenced by x and p.
// Returns the task index.
func receiveInput(addr string, x, p interface{}) (int, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return 0, errors.New("connect to server: " + err.Error())
	}
	defer conn.Close()

	// Send (empty) input request.
	req := inputReq()
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return 0, errors.New("send input request: " + err.Error())
	}

	// Decode response.
	resp := &inputResp{X: x, P: p}
	if err := json.NewDecoder(conn).Decode(resp); err != nil {
		return 0, errors.New("decode input response: " + err.Error())
	}
	return resp.Index, nil
}

func sendOutput(addr string, index int, y interface{}, taskerr error) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return errors.New("connect to server: " + err.Error())
	}
	defer conn.Close()

	req := &outputReq{index, y, taskerr}
	if err := json.NewEncoder(conn).Encode(req.Generic()); err != nil {
		return errors.New("send output request: " + err.Error())
	}

	// No data to receive. Done.
	return nil
}
