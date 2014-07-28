package dstrfn

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
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
	dir := os.Getenv("PBS_O_WORKDIR")
	if len(dir) == 0 {
		panic("environment variable empty: PBS_O_WORKDIR")
	}
	if err := os.Chdir(dir); err != nil {
		panic(fmt.Sprintf("chdir: %v", err))
	}

	// Request input from the master.
	xptr := task.NewInput()
	p := task.NewConfig()
	log.Println("receive input")
	index, err := receiveInput(addrStr, xptr, p)
	if err != nil {
		panic(err)
	}

	x := reflect.ValueOf(xptr).Elem().Interface()
	log.Println("call function")
	y, taskerr := task.Func(x, p)

	log.Println("send output")
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

	req := &outputReq{index, y, errToStr(taskerr)}
	if err := json.NewEncoder(conn).Encode(req.Generic()); err != nil {
		return errors.New("send output request: " + err.Error())
	}

	// No data to receive. Done.
	return nil
}

// Assumes no difference between nil error and empty string
// for serialization purposes.
func errToStr(err error) *string {
	if err == nil {
		return nil
	}
	s := err.Error()
	return &s
}
