package main

import (
	"flag"
	"fmt"
	"github.com/jackvalmadre/go-grideng"
	"io"
	"log"
)

// Input is an integer.
type Input int

// Input has a unique identifier.
func (input Input) Name() string { return fmt.Sprint(int(input)) }

// Output is an integer.
type Output int

// Task is defined by an input and an execute method.
type Task struct{ X Input }

func (task Task) Input() grideng.Input { return task.X }

// Map x to x squared.
func (task Task) Execute() (grideng.Output, error) {
	x := int(task.X)
	y := x * x
	return Output(y), nil
}

func main() {
	slave := flag.Bool("slave", false, "Run in slave mode?")
	n := flag.Int("n", 32, "Number of jobs")
	flag.Parse()

	if *slave {
		// Get Grid Engine task ID.
		num, err := grideng.TaskNumFromEnv("SGE_TASK_ID")
		if err != nil {
			panic(err)
		}
		// Be a slave.
		grideng.Slave(num, InputReader{})
		return
	}

	// Populate inputs.
	x := make([]int, *n)
	for i := range x {
		x[i] = i + 1
	}
	// Execute all tasks.
	files, err := grideng.Master(InputList(x), "", []string{"-slave"})
	if err != nil {
		log.Fatal(err)
	}
	// Load outputs.
	y := make([]int, *n)
	grideng.LoadOutputs(OutputList(y), files)

	for i := range y {
		fmt.Printf("%6d: %6d -> %6d\n", i, x[i], y[i])
	}
	// Output:
	//      0:      1 ->      1
	//      1:      2 ->      4
	//      2:      3 ->      9
	// ...
	//     31:     32 ->   1024
}

type InputList []int

func (list InputList) Len() int               { return len(list) }
func (list InputList) At(i int) grideng.Input { return Input(list[i]) }

// How to write an input.
func (input Input) Write(w io.Writer) error {
	if _, err := fmt.Fprintln(w, int(input)); err != nil {
		return err
	}
	return nil
}

type InputReader struct{}

// How to read an input (into a task).
func (reader InputReader) Read(r io.Reader) (grideng.Task, error) {
	x, err := readInt(r)
	if err != nil {
		return nil, err
	}
	input := Input(x)
	task := Task{input}
	return task, nil
}

// How to write an output.
func (output Output) Write(w io.Writer) error {
	if _, err := fmt.Fprintln(w, int(output)); err != nil {
		return err
	}
	return nil
}

type OutputList []int

// How to read an output (into a list).
func (list OutputList) Read(i int, r io.Reader) error {
	x, err := readInt(r)
	if err != nil {
		return err
	}
	list[i] = x
	return nil
}

// Reads single integer from file.
func readInt(r io.Reader) (int, error) {
	var x int
	_, err := fmt.Fscanln(r, &x)
	return x, err
}
