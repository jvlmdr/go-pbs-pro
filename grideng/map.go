package grideng

import (
	"fmt"
	"io"
	"os"
)

var DefaultStdout io.Writer = os.Stdout
var DefaultStderr io.Writer = os.Stderr

// Computes y[i] = f(x[i], p) for all i.
//
// Parameters:
// 	name identifies the registered task f.
// 	y is a slice to which the outputs will be assigned.
//	x is a slice of inputs.
// 	p contains optional constant parameters to f.
func MapWriteTo(name string, y, x, p interface{}, stdout, stderr io.Writer) error {
	task, there := tasks[name]
	if !there {
		return fmt.Errorf(`task not found: "%s"`, name)
	}
	return master(task, name, y, x, p, stdout, stderr)
}

// Calls MapWriteTo with DefaultStdout and DefaultStderr.
func Map(name string, y, x, p interface{}) error {
	return MapWriteTo(name, y, x, p, DefaultStdout, DefaultStderr)
}
