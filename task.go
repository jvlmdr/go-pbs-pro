package grideng

import "io"

// Reads a task input (e.g. from file).
type InputReader interface {
	// Returns task input.
	Read(io.Reader) (Input, error)
}

// Input type which also defines the task.
type Input interface {
	// Writes task input to file, read by InputReader.Read().
	Write(io.Writer) error
	// Uniquely identifying name.
	Name() string
	// Turn the input into a task.
	Execute() (Output, error)
}

type Output interface {
	// Writes task output to file.
	Write(io.Writer) error
}
