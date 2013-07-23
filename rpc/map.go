package main

type Map interface {
	Len() int
	// Returns a task which provides access to its input and output.
	Task(i int) Task
	// Returns the i-th input.
	Input(i int) interface{}
	// Returns a pointer to the i-th output, suitable for use with Decode.
	Output(i int) interface{}
}
