package grideng

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
)

// Panics if x is not a slice.
// y should be the same length as x.
func master(task *qsubTask, name string, y, x, p interface{}) error {
	n := reflect.ValueOf(x).Len()

	// Open port for server.
	l, err := net.Listen("tcp", addrStr)
	if err != nil {
		return err
	}
	defer l.Close()

	// Start server.
	todo := make(chan int)
	go func(n int) {
		// Thread-safely obtain task IDs.
		for i := 0; i < n; i++ {
			todo <- i
		}
	}(n)
	dsts := make(chan interface{})
	go func(n int) {
		// Thread-safely call Task.NewOutput().
		for i := 0; i < n; i++ {
			dsts <- task.Task.NewOutput()
		}
	}(n)
	errs := make(chan error)
	go serve(l, task.Task, name, y, x, p, todo, dsts, errs)

	// Submit job.
	var args []string
	args = append(args, "-grideng.task", name)
	args = append(args, "-grideng.addr", addrStr)
	proc := make(chan error)
	go func() {
		proc <- submit(n, *task.Res, args)
	}()

	// Wait for all tasks to finish.
	// Do not exit if one task fails.
	var (
		num   int
		first error
		exit  bool
	)
	for num < n && !exit {
		select {
		case err := <-errs:
			if err != nil && first != nil {
				first = err
			}
			n++
		case err := <-proc:
			log.Println("qsub exit")
			if err != nil {
				return err
			}
			exit = true
		}
	}
	if first != nil {
		return first
	}
	return nil
}

// y is a slice of destinations.
// p is a configuration object, possibly nil.
// x is a slice of inputs.
func serve(l net.Listener, task Task, name string, y, x, p interface{}, todo <-chan int, dsts <-chan interface{}, errs chan<- error) {
	for {
		conn, err := l.Accept()
		// The listener will be closed when qsub exits.
		if err != nil {
			log.Println("accept:", err)
			break
		}

		go func(conn net.Conn) {
			defer conn.Close()
			handle(conn, y, x, p, todo, dsts, errs)
		}(conn)
	}
}

func handle(conn io.ReadWriter, y, x, p interface{}, todo <-chan int, dsts <-chan interface{}, errs chan<- error) {
	// Read request.
	req := new(request)

	dec := json.NewDecoder(conn)
	if err := dec.Decode(req); err != nil {
		log.Fatalln("decode request:", err)
	}

	switch req.Type {
	default:
		err := fmt.Errorf(`unknown request type: "%s"`, req.Type)
		panic(err)

	case recvType:
		i := <-todo
		xi := reflect.ValueOf(x).Index(i).Interface()
		resp := &inputResp{i, xi, p}
		if err := json.NewEncoder(conn).Encode(resp); err != nil {
			log.Fatalln("send input:", err)
		}

	case sendType:
		body := &outputReq{Y: <-dsts}
		if err := json.Unmarshal(req.Body, body); err != nil {
			log.Fatalln("decode output request:", err)
		}
		// Send the error if one occurred, nil otherwise.
		if body.Err != nil {
			errs <- body.Err
			return
		}
		// Assign value to output slice.
		reflect.ValueOf(y).Index(body.Index).Set(reflect.ValueOf(body.Y).Elem())
		errs <- nil
	}
}