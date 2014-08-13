package dstrfn

import (
	"fmt"
	"reflect"
)

// Defines a map task from a function.
//
// The function must have at least one input.
// It will be evaluated with the first argument taking every value in a list.
// The remaining arguments are held constant.
// The function must have either one or two outputs.
// The second output, if present, must be an error.
// The inputs and the first output must be concrete types
// suitable for decoding into.
//
// Examples:
//	var (
//		sqr   = dstrfn.ConfigFunc(func(x float64) float64 { return x * x })
//		sqrt  = dstrfn.ConfigFunc(math.Sqrt)
//		pow   = dstrfn.ConfigFunc(math.Pow)
//		linop = dstrfn.ConfigFunc(func(x *Vec, a *Mat) float64 { return a.Mul(x) })
//	)
//
// The key difference between Map and Func is that Func does not use Config().
// Currently ConfigFunc must have a non-error return value.
func ConfigFunc(f interface{}) ConfigTask {
	ftyp := reflect.TypeOf(f)
	if ftyp.Kind() != reflect.Func {
		panic(fmt.Sprintf("not func: %v", ftyp.Kind()))
	}
	if n := ftyp.NumIn(); n == 0 {
		panic("expect at least one input")
	}
	if n := ftyp.NumOut(); n == 0 {
		panic("expect at least one output")
	} else if n > 2 {
		panic(fmt.Sprintf("more than two outputs: %d", n))
	} else if n == 2 {
		errtyp := ftyp.Out(1)
		if isError(errtyp) {
			panic(fmt.Sprintf("output type is not error: %v", errtyp))
		}
	}
	return &mapTask{f}
}

// Task defined by a function.
type mapTask struct {
	F interface{}
}

// Returns a new object of the type of the first argument.
func (t *mapTask) NewInput() interface{} {
	ftyp := reflect.TypeOf(t.F)
	return reflect.New(ftyp.In(0)).Interface()
}

// Creates a list of new objects with the types of the remaining arguments.
// Returns nil if there is only one argument.
func (t *mapTask) NewConfig() interface{} {
	ftyp := reflect.TypeOf(t.F)
	n := ftyp.NumIn() - 1
	if n == 0 {
		return nil
	}
	in := make([]interface{}, n)
	for i := range in {
		in[i] = reflect.New(ftyp.In(i + 1)).Interface()
	}
	return &in
}

// Returns a new object of the type of the first return value.
func (t *mapTask) NewOutput() interface{} {
	ftyp := reflect.TypeOf(t.F)
	return reflect.New(ftyp.Out(0)).Interface()
}

// If function only takes one argument then p is ignored.
func (t *mapTask) Func(x, p interface{}) (interface{}, error) {
	fval := reflect.ValueOf(t.F)
	ftyp := fval.Type()
	in := []reflect.Value{reflect.ValueOf(x)}
	// Append additional arguments if there are any.
	if ftyp.NumIn() > 1 {
		args := p.([]interface{})
		for _, arg := range args {
			// De-reference each element.
			in = append(in, reflect.ValueOf(arg).Elem())
		}
	}
	// Panics if call is invalid.
	out := reflect.ValueOf(t.F).Call(in)
	// Panics if f has no return values.
	if t.HasError() {
		err := out[1].Interface()
		if err != nil {
			// Panics if second return value is not assignable to error.
			return nil, err.(error)
		}
	}
	return out[0].Interface(), nil
}

func (t *mapTask) HasError() bool {
	ftyp := reflect.TypeOf(t.F)
	return ftyp.NumOut() > 1
}
