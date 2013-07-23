package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
)

// A request to send output contains an index, and possibly an error (plus the body).
type OutputRequestHeader struct {
	Index int
	Error string
}

// A response to send input contains the index (plus the body).
type InputResponseHeader struct {
	Index int
}

type SquareMap struct {
	X []float64
}

func (m SquareMap) Len() int                 { return len(m.X) }
func (m SquareMap) Input(i int) interface{}  { return m.X[i] }
func (m SquareMap) Output(i int) interface{} { return &m.X[i] }

func (m SquareMap) Task(i int) Task {
	return &SquareTask{m.X[i]}
}

type SquareTask struct{ X float64 }

func (t *SquareTask) Input() interface{} { return &t.X }
func (t SquareTask) Output() interface{} { return t.X }

func (t *SquareTask) Do() error {
	// This is the actual work.
	t.X = t.X * t.X
	return nil
}

func main() {
	master := flag.Bool("master", false, "Operate in master mode?")
	slave := flag.Bool("slave", false, "Operate in slave mode?")
	addr := flag.String("addr", "", "Master address")
	port := flag.Int("port", 1234, "Master port")
	n := flag.Int("n", 0, "Number of jobs")
	flag.Parse()

	if *slave && !*master {
		task := new(SquareTask)
		Slave(task, *addr, *port)
		return
	}
	if *slave == *master {
		fmt.Fprintln(os.Stderr, "Must specify master xor slave")
		os.Exit(1)
	}
	Master(*n, *addr, *port)
	return
}

func Slave(task Task, addr string, port int) {
	var index int

	func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
		if err != nil {
			log.Fatalln("Dial (to send output) error:", err)
		}
		defer conn.Close()

		codec := MakeJSONClientCodec(conn)
		// Send request.
		if err := codec.WriteRequest(InputRequest, nil, nil); err != nil {
			log.Fatalln("Write request (to receive input) error:", err)
		}

		// Read response type.
		response, err := codec.ReadResponse()
		if err != nil {
			log.Fatalln("Read response error:", err)
		}
		// Read response header.
		var header InputResponseHeader
		if err := response.ReadHeader(&header); err != nil {
			log.Fatalln("Read response (to receive input) header error:", err)
		}
		index = header.Index
		log.Println("Task index:", index)
		// Read response body.
		if err := response.ReadBody(task.Input()); err != nil {
			log.Fatalln("Read response (to receive input) body error:", err)
		}
	}()

	if err := task.Do(); err != nil {
		log.Fatalln("Do task error:", err)
	}

	func() {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
		if err != nil {
			log.Fatalln("Dial (to send output) error:", err)
		}
		defer conn.Close()

		// Prepare header.
		header := OutputRequestHeader{Index: index, Error: ""}

		codec := MakeJSONClientCodec(conn)
		// Send request.
		if err := codec.WriteRequest(OutputRequest, header, task.Output()); err != nil {
			log.Fatalln("Write request (to send output) error:", err)
		}
		// No data to receive. Done.
	}()
}

func Master(n int, addr string, port int) {
	if n <= 0 {
		fmt.Fprintln(os.Stderr, "Require n > 0")
		os.Exit(1)
	}

	// Initialize inputs.
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = float64(i)
	}

	var m Map = SquareMap{x}

	// Start server.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalln("listen error:", err)
	}

	in := make(chan int)
	go func() {
		for i := 0; i < n; i++ {
			in <- i
		}
	}()

	// Start server. Listens for both results and requests.
	go func() {
		for i := 0; i < 2*n; i++ {
			// Block on receiving connection.
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalln("Listen error:", err)
			}

			// Handle connection.
			go func() {
				defer conn.Close()

				// Read request type.
				codec := MakeJSONServerCodec(conn)
				request, err := codec.ReadRequest()
				if err != nil {
					log.Fatalln("Read request type error:", err)
				}

				switch request.Type() {
				default:
					log.Fatalf(`Unknown request type "%s"`, request.Type())
				case InputRequest:
					// No more information to read since input request has no content.
					// Get index of job.
					index := <-in
					// Prepare header and send input.
					header := InputResponseHeader{index}
					if err := codec.WriteResponse(header, m.Input(index)); err != nil {
						log.Fatalln("Write response (to receive input) error:", err)
					}
				case OutputRequest:
					// Read header and content of request to send output.
					var header OutputRequestHeader
					if err := request.ReadHeader(&header); err != nil {
						log.Fatalln("Read request (to send output) header error:", err)
					}
					index := header.Index
					// Read body into map output.
					if err := request.ReadBody(m.Output(index)); err != nil {
						log.Fatalln("Read request (to send output) body error:", err)
					}
					// No need to send response to client.
				}
			}()
		}
	}()

	// Start qsub.
	args := []string{"-slave", fmt.Sprintf("-addr=%s", addr), fmt.Sprintf("-port=%d", port)}
	err = Submit(n, "", args)
	if err != nil {
		log.Fatalln("submit error:", err)
	}

	fmt.Println(x)
}

func Submit(n int, resources string, cmdArgs []string) error {
	var args []string
	// Submitting a binary job.
	args = append(args, "-b", "y")
	// Wait for jobs to finish.
	args = append(args, "-sync", "y")
	// Use current working directory.
	args = append(args, "-cwd")
	// Use same environment variables.
	args = append(args, "-V")
	// Use same environment variables.
	args = append(args, "-t", fmt.Sprintf("1-%d", n))
	// Set resources.
	if len(resources) > 0 {
		args = append(args, "-l", resources)
	}
	// Redirect stdout.
	args = append(args, "-o", `stdout-$TASK_ID`)
	// Redirect stderr.
	args = append(args, "-e", `stderr-$TASK_ID`)

	// Name of executable to run.
	args = append(args, os.Args[0])
	args = append(args, cmdArgs...)

	// Submit.
	cmd := exec.Command("qsub", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Print("qsub")
	for _, arg := range args {
		fmt.Print(" ", arg)
	}
	fmt.Println()

	return cmd.Run()
}
