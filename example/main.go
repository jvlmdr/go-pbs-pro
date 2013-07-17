package main

import (
	"flag"
	"fmt"
	"github.com/jackvalmadre/go-grideng"
	"io"
	"log"
)

func main() {
	// Operate in master or slave mode?
	master := flag.Bool("master", false, "Run in master mode?")
	slave := flag.Bool("slave", false, "Run in slave mode?")
	// Number of jobs to run.
	n := flag.Int("master.n", 32, "Number of jobs")
	resources := flag.String("master.l", "", "qsub resources string")
	flag.Parse()

	if !*master && !*slave {
		log.Fatalf("Master/slave mode not specified")
	}

	if *slave {
		num, err := grideng.TaskNumFromEnv("SGE_TASK_ID")
		if err != nil {
			panic(err)
		}

		grideng.Slave(num, InputReader{})
		return
	}

	// Populate tasks.
	inputs := make([]grideng.Input, *n)
	for i := range inputs {
		inputs[i] = Input(i + 1)
	}

	args := []string{"-slave"}
	results, err := grideng.Master(inputs, *resources, args)
	if err != nil {
		log.Fatal(err)
	}

	for name, filename := range results {
		fmt.Println(name, filename)
	}
}

type Input int

// Unique identifier.
func (input Input) Name() string {
	return fmt.Sprintf("%010d", int(input))
}

// Write input to disk.
func (input Input) Write(w io.Writer) error {
	if _, err := fmt.Fprintln(w, int(input)); err != nil {
		return err
	}
	return nil
}

// Read input from disk.
func readInput(r io.Reader) (Input, error) {
	var x int
	_, err := fmt.Fscanln(r, &x)
	if err != nil {
		return Input(0), err
	}
	return Input(x), nil
}

type Output int

type Task Input

func (task Task) Input() grideng.Input {
	return Input(task)
}

func (task Task) Execute() (grideng.Output, error) {
	n := int(Input(task))
	output := Output(n * n)
	return output, nil
}

// Write output to disk.
func (output Output) Write(w io.Writer) error {
	if _, err := fmt.Fprintln(w, int(output)); err != nil {
		return err
	}
	return nil
}

type InputReader struct{}

func (reader InputReader) Read(r io.Reader) (grideng.Task, error) {
	input, err := readInput(r)
	if err != nil {
		return nil, err
	}

	task := Task(input)
	return task, nil
}
