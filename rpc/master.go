package rpc

import (
	"fmt"
	"log"
	"net"
)

// Always attempts every task, even if one returns an error.
// Returns an error if any task returns an error.
func Do(m Map, addr, port, codec, resources string) error {
	// Queue up task indices.
	todo := make(chan int)
	go countTo(m.Len(), todo)

	// When the server receives a task's result, it is sent down this channel.
	// If the task succeeds, a nil error is sent.
	errs := make(chan error)

	// If something gets sent along this channel, it means that all tasks communicated a result.
	// If all tasks succeed, nil will be passed along the channel.
	done := make(chan error, 1)
	go func() { done <- receiveResults(m.Len(), errs) }()

	// Send result of qsub down this channel when it exits.
	qsub := make(chan error, 1)

	// Prepare to start qsub.
	var args []string
	args = append(args, "-slave")
	args = append(args, fmt.Sprintf("-addr=%s", addr))
	args = append(args, fmt.Sprintf("-port=%s", port))
	args = append(args, fmt.Sprintf("-codec=%s", codec))

	// Open port.
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}
	// Start listening for requests.
	go serveRequests(m, l, codec, todo, errs)
	// Call qsub and close the port when it finishes.
	go func() {
		// Keep port open until qsub closes.
		defer l.Close()
		// Execute qsub synchronously.
		err := submit(m.Len(), resources, args)
		log.Println("qsub finished")
		qsub <- err
	}()

	// Wait for either qsub to exit or all tasks to finish.
	// Buffer both channels so that if we return for another reason, the routines can still end.
	for {
		select {
		case err := <-done:
			// All jobs communicated a result.
			log.Println("All jobs finished")
			return err
		case err := <-qsub:
			// If qsub returned an error, probably means some tasks couldn't communicate.
			if err != nil {
				return err
			}
			// If qsub did not return an error, wait for tasks to finish.
			// This could result in a hang!
			log.Println("qsub exited without incident")
		}
	}
}

// Accept() will trigger an error when the listener is closed.
// This error should be ignored if all jobs have finished, therefore the channel buffer size should be at least 1.
func serveRequests(m Map, l net.Listener, codec string, todo <-chan int, errs chan<- error) {
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer conn.Close()
			handle(m, conn, codec, todo, errs)
		}()
	}
}

// Sends the numbers 0, 1, ..., n-1 down the channel.
func countTo(n int, ch chan<- int) {
	for i := 0; i < n; i++ {
		ch <- i
	}
}

// Receives n errors (nil or otherwise) before returning first non-nil error.
func receiveResults(n int, errs <-chan error) error {
	var err error
	for i := 0; i < n; i++ {
		next := <-errs
		if err == nil {
			err = next
		}
	}
	return err
}
