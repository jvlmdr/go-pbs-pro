package grideng

import (
	"fmt"
	"reflect"
)

// Describes a task to execute on a remote host.
//
// Captures the minimum required by master and slave.
//
// NewInput(), NewConfig() and Func() are called by the slave,
// NewOutput() is called by the master.
type Task interface {
	// Returns an input object which can be decoded into.
	//	x := task.NewInput()
	//	json.NewDecoder(r).Decode(x)
	// This will be de-referenced before passing to Func().
	NewInput() interface{}
	// Returns a config object which can be decoded into,
	// or nil if there are no config parameters.
	NewConfig() interface{}

	// Performs the task given input x and parameters p.
	Func(x, p interface{}) (interface{}, error)

	// Returns an output object which can be decoded into.
	//	y := task.NewOutput()
	//	json.NewDecoder(r).Decode(y)
	NewOutput() interface{}
}

// Defines a map task from a function.
//
// The function f must take either one or two arguments and
// have either one or two return values.
// The second return value must be assignable to error.
// The argument and first return value must be concrete types
// for use with reflect.New().
//
// Examples:
//	sqr := grideng.Func(func(x float64) float64 { return x * x })
//	sqrt := grideng.Func(math.Sqrt)
//	pow := grideng.Func(math.Pow)
//	linop := grideng.Func(func(x *Vec, a *Mat) float64 { a.Mul(x) })
func Func(f interface{}) Task {
	fval := reflect.ValueOf(f)
	if fval.Kind() != reflect.Func {
		panic(fmt.Sprintf("not func: %v", fval.Kind()))
	}
	if n := fval.Type().NumIn(); n < 1 || n > 2 {
		panic(fmt.Sprintf("number of arguments: %d", n))
	}
	if n := fval.Type().NumOut(); n < 1 || n > 2 {
		panic(fmt.Sprintf("number of return values: %d", n))
	}
	return &funcTask{f}
}

// Task defined by a function.
type funcTask struct {
	F interface{}
}

// Returns a new object of the type of the first argument.
func (t *funcTask) NewInput() interface{} {
	f := reflect.ValueOf(t.F)
	return reflect.New(f.Type().In(0)).Interface()
}

// Returns a new object of the type of the second argument.
// Returns nil if there is no second argument.
func (t *funcTask) NewConfig() interface{} {
	f := reflect.ValueOf(t.F)
	if f.Type().NumIn() < 2 {
		return nil
	}
	return reflect.New(f.Type().In(1)).Interface()
}

// Returns a new object of the type of the first return value.
func (t *funcTask) NewOutput() interface{} {
	f := reflect.ValueOf(t.F)
	return reflect.New(f.Type().Out(0)).Interface()
}

// If function only takes one argument then p is ignored.
func (t *funcTask) Func(x, p interface{}) (interface{}, error) {
	f := reflect.ValueOf(t.F)
	in := []reflect.Value{reflect.ValueOf(x)}
	// Only use second argument if function accepts one.
	if f.Type().NumIn() > 1 {
		in = append(in, reflect.ValueOf(p).Elem())
	}
	// Panics if call is invalid.
	out := f.Call(in)
	// Panics if f has no return values.
	y := out[0].Interface()
	if len(out) == 1 {
		return y, nil
	}
	err := out[1].Interface()
	if err == nil {
		return y, nil
	}
	// Panics if second return value is not assignable to error.
	return y, err.(error)
}
