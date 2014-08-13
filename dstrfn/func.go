package dstrfn

import (
	"fmt"
	"reflect"
)

// Func creates a task from a function.
// The function can have up to two outputs.
// The last output to have type error will be treated as an error.
// All inputs and all other outputs must have concrete types
// suitable for use with reflect.New().
func Func(f interface{}) Task {
	ftyp := reflect.TypeOf(f)
	if ftyp.Kind() != reflect.Func {
		panic(fmt.Sprintf("not func: %v", ftyp.Kind()))
	}
	n := ftyp.NumOut()
	if n > 0 {
		// Remove last type if it is error.
		if isError(ftyp.Out(n - 1)) {
			n--
		}
		if n > 1 {
			panic(fmt.Sprintf("more than one non-error output: %d", n))
		}
	}
	return &funcTask{f}
}

type funcTask struct {
	F interface{}
}

func (t *funcTask) NewInput() interface{} {
	ftyp := reflect.TypeOf(t.F)
	n := ftyp.NumIn()
	if n == 0 {
		panic("function must have at least one input")
	}
	if n == 1 {
		return reflect.New(ftyp.In(0)).Interface()
	}
	// Multiple input arguments.
	in := make([]interface{}, n)
	for i := range in {
		in[i] = reflect.New(ftyp.In(i)).Interface()
	}
	return &in
}

func (t *funcTask) NewOutput() interface{} {
	if !t.HasOutput() {
		return nil
	}
	return reflect.New(reflect.TypeOf(t.F).Out(0)).Interface()
}

func (t *funcTask) Func(x interface{}) (interface{}, error) {
	fval := reflect.ValueOf(t.F)
	ftyp := fval.Type()
	if ftyp.NumIn() == 0 {
		panic("function must have at least one input")
	}
	var in []reflect.Value
	if ftyp.NumIn() == 1 {
		// If there is only one argument, use it directly.
		in = append(in, reflect.ValueOf(x))
	} else {
		// If there are multiple arguments, convert to []interface{}.
		// Derference each element.
		args := x.([]interface{})
		for _, arg := range args {
			in = append(in, reflect.ValueOf(arg).Elem())
		}
	}
	// Panics if call is invalid.
	out := reflect.ValueOf(t.F).Call(in)

	if t.HasError() {
		// Convert last output to error.
		err := out[len(out)-1].Interface()
		if err != nil {
			return nil, err.(error)
		}
	}
	if !t.HasOutput() {
		return nil, nil
	}
	return out[0].Interface(), nil
}

func (t *funcTask) HasOutput() bool {
	ftyp := reflect.TypeOf(t.F)
	n := ftyp.NumOut()
	if n == 0 {
		return false
	}
	if n > 1 {
		return true
	}
	return !isError(ftyp.Out(0))
}

func (t *funcTask) HasError() bool {
	ftyp := reflect.TypeOf(t.F)
	n := ftyp.NumOut()
	return isError(ftyp.Out(n - 1))
}
