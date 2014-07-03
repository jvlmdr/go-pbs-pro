package grideng

import (
	"fmt"
	"io"
	"os"
)

var DefaultCmdOut io.Writer = os.Stdout
var DefaultCmdErr io.Writer = os.Stderr

var DefaultJobOut string = "/dev/null"
var DefaultJobErr string = "/dev/null"

// Computes y[i] = f(x[i], p) for all i.
//
// Parameters:
// 	name identifies the registered task f.
// 	y is a slice to which the outputs will be assigned.
//	x is a slice of inputs.
// 	p contains optional constant parameters to f.
func MapWriteTo(name string, y, x, p interface{}, cmdout, cmderr io.Writer, jobout, joberr string) error {
	task, there := tasks[name]
	if !there {
		return fmt.Errorf(`task not found: "%s"`, name)
	}
	return master(task, name, y, x, p, cmdout, cmderr, jobout, joberr)
}

// Calls MapWriteTo with defaults.
func Map(name string, y, x, p interface{}) error {
	return MapWriteTo(name, y, x, p, DefaultCmdOut, DefaultCmdErr, DefaultJobOut, DefaultJobErr)
}
