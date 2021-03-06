package dstrfn

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// Default place to route stdout and stderr of qsub when invoked.
var (
	DefaultCmdOut io.Writer = os.Stdout
	DefaultCmdErr io.Writer = os.Stderr
)

// Map computes y[i] = f(x[i], p) for all i.
//
// The function is specified by name and must already be registered.
// The input x must be a slice or similar (see reflect.Value.Index()).
// The output y must be a pointer to a slice.
// If the length of y is sufficient to hold the output, it will be over-written.
// If it is not sufficient, a new array will be allocated.
// After a succesful call, the length of y will match that of x.
func Map(f string, y, x, p interface{}) error {
	return MapWriteTo(f, y, x, p, DefaultCmdOut, DefaultCmdErr)
}

func MapWriteTo(f string, y, x, p interface{}, cmdout, cmderr io.Writer) error {
	task, there := tasks[f]
	if !there {
		return fmt.Errorf(`task not found: "%s"`, f)
	}

	n := reflect.ValueOf(x).Len()
	// This changes the type of y from *[]Y to []Y.
	y = ensureLenAndDeref(y, n)

	var u, v interface{}
	if task.Chunk {
		m := max(task.ChunkLen, 1)
		u = split(x, 1, m)
		l := reflect.ValueOf(u).Len()
		v = reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(y)), l, l).Interface()
	} else {
		u = x
		v = y
	}

	userargs := strings.Split(task.Flags, " ")
	err := master(task.Task, f, v, u, p, userargs, cmdout, cmderr, task.Stdout, task.Stderr)
	if err != nil {
		return err
	}

	if task.Chunk {
		reflect.Copy(reflect.ValueOf(y), reflect.ValueOf(merge(v)))
	}
	return nil
}

// Ensures that dst has length n and then de-references the pointer.
// The slice header is sufficient to change the underlying elements.
func ensureLenAndDeref(dst interface{}, n int) interface{} {
	// De-reference pointer.
	val := reflect.ValueOf(dst).Elem()
	if val.Len() >= n {
		return val.Slice(0, n).Interface()
	}
	// Not big enough, re-allocate.
	val.Set(reflect.MakeSlice(val.Type(), n, n))
	return val.Interface()
}
