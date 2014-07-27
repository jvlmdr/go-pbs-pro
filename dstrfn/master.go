package dstrfn

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"time"
)

func listenRetry(netstr, laddr string) net.Listener {
	for {
		// Open port for server.
		l, err := net.Listen(netstr, laddr)
		if err == nil {
			return l
		}
		log.Println(err)
		// Pause.
		time.Sleep(time.Second)
		log.Println("listen: try again")
	}
}

// Panics if x is not a slice.
// y should be the same length as x.
func master(task Task, name string, y, x, p interface{}, res string, cmdout, cmderr io.Writer, jobout, joberr string) error {
	n := reflect.ValueOf(x).Len()

	// Open port for server.
	l := listenRetry("tcp", addrStr)
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
			dsts <- task.NewOutput()
		}
	}(n)
	errs := make(chan error)
	go serve(l, task, name, y, x, p, todo, dsts, errs)

	// Submit job.
	var args []string
	args = append(args, "-dstrfn.task", name)
	args = append(args, "-dstrfn.addr", addrStr)
	proc := make(chan error)
	go func() {
		proc <- submit(n, res, args, cmdout, cmderr, jobout, joberr)
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
			if err != nil && first == nil {
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
// x is a slice of inputs.
// p is a configuration object, possibly nil.
func serve(l net.Listener, task Task, name string, y, x, p interface{}, todo <-chan int, dsts <-chan interface{}, errs chan<- error) {
	for {
		conn, err := l.Accept()
		// The listener will be closed when qsub exits.
		if err != nil {
			log.Println("accept:", err)
			break
		}

		go func(conn net.Conn) {
			err := handleClose(conn, y, x, p, todo, dsts)
			errs <- err
		}(conn)
	}
}

// Ensures the connection is closed before sending result down the channel.
// Catches any errors that occur in conn.Close().
func handleClose(conn net.Conn, y, x, p interface{}, todo <-chan int, dsts <-chan interface{}) error {
	handleErr := handle(conn, y, x, p, todo, dsts)
	closeErr := conn.Close()
	if handleErr != nil {
		return handleErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

func handle(rw io.ReadWriter, y, x, p interface{}, todo <-chan int, dsts <-chan interface{}) error {
	// Read request.
	req := new(request)
	if err := json.NewDecoder(rw).Decode(req); err != nil {
		return fmt.Errorf("receive request: %v", err)
	}

	switch req.Type {
	default:
		// Error occurred in protocol, not user code.
		return fmt.Errorf(`unknown request type: "%s"`, req.Type)

	case recvType:
		i := <-todo
		xi := reflect.ValueOf(x).Index(i).Interface()
		resp := &inputResp{i, xi, p}
		if err := json.NewEncoder(rw).Encode(resp); err != nil {
			return fmt.Errorf("send input: %v", err)
		}
		return nil

	case sendType:
		body := &outputReq{Y: <-dsts}
		if err := json.Unmarshal(req.Body, body); err != nil {
			return fmt.Errorf("receive output: %v", err)
		}
		// Send the error if one occurred, nil otherwise.
		if body.Err != nil {
			return fmt.Errorf("slave error: %s", *body.Err)
		}
		// Assign value to output slice.
		reflect.ValueOf(y).Index(body.Index).Set(reflect.ValueOf(body.Y).Elem())
		return nil
	}
}
