package grideng

import "reflect"

// Describes a task to execute on a remote host.
//
// Should capture the minimum required by master and slave.
//
// NewInput(), NewConfig() and Func() are called by the slave,
// NewOutput() is called by the master.
type Task interface {
	// Returns an input object which can be decoded into.
	//	x := task.NewInput()
	//	json.NewDecoder(r).Decode(x)
	NewInput() interface{}
	// Returns a config object which can be decoded into,
	// or nil if there are no config parameters.
	NewConfig() interface{}

	// Performs the task.
	Func(x, p interface{}) (interface{}, error)

	// Returns an output object which can be decoded into.
	//	y := task.NewOutput()
	//	json.NewDecoder(r).Decode(y)
	NewOutput() interface{}
}

// Task defined by a function.
//
// The function must take either one or two arguments and
// have either one or two return values.
// The second return value must have type error.
// The argument and first return value must be concrete types
// for use with reflect.New().
type Func struct {
	F interface{}
}

// Returns a new object of the type of the first argument.
func (t *Func) NewInput() interface{} {
	f := reflect.ValueOf(t.F)
	return reflect.New(f.Type().In(0)).Interface()
}

// Returns a new object of the type of the second argument.
// Returns nil if there is no second argument.
func (t *Func) NewConfig() interface{} {
	f := reflect.ValueOf(t.F)
	if f.Type().NumIn() < 2 {
		return nil
	}
	return reflect.New(f.Type().In(1)).Interface()
}

// Returns a new object of the type of the first return value.
func (t *Func) NewOutput() interface{} {
	f := reflect.ValueOf(t.F)
	return reflect.New(f.Type().Out(0)).Interface()
}

// If function only takes one argument then p is ignored.
func (t *Func) Func(x, p interface{}) (interface{}, error) {
	f := reflect.ValueOf(t.F)
	in := []reflect.Value{reflect.ValueOf(x).Elem()}
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
	// Panics if second return value is not an error.
	err := out[1].Interface().(error)
	// Ignore any further return values.
	return y, err
}
