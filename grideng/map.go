package grideng

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

var DefaultCmdOut io.Writer = os.Stdout
var DefaultCmdErr io.Writer = os.Stderr

// Computes y[i] = f(x[i], p) for all i.
//
// Parameters:
// 	name identifies the registered task f.
// 	y is a slice to which the outputs will be assigned.
//	x is a slice of inputs.
// 	p contains optional constant parameters to f.
func MapWriteTo(name string, y, x, p interface{}, cmdout, cmderr io.Writer) error {
	task, there := tasks[name]
	if !there {
		return fmt.Errorf(`task not found: "%s"`, name)
	}

	m := max(task.ChunkLen, 1)
	u := split(x, 1, m)
	v := split(y, 1, m)
	err := master(task.Task, name, v, u, p, task.Res, cmdout, cmderr, task.Stdout, task.Stderr)
	if err != nil {
		return err
	}
	reflect.Copy(reflect.ValueOf(y), reflect.ValueOf(merge(v)))
	return nil
}

// Calls MapWriteTo with defaults.
func Map(name string, y, x, p interface{}) error {
	return MapWriteTo(name, y, x, p, DefaultCmdOut, DefaultCmdErr)
}
