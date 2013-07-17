package grideng

import "io"

type Input interface {
	// Uniquely defines an input.
	Name() string
	Write(io.Writer) error
}

type Task interface {
	Input() Input
	Execute() (Output, error)
}

type Output interface {
	Write(io.Writer) error
}

type InputReader interface {
	Read(io.Reader) (Task, error)
}

type OutputReader interface {
	Read(io.Reader) (Output, error)
}
