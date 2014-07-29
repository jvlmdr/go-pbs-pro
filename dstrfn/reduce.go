package dstrfn

import (
	"fmt"
	"io"
	"reflect"
)

// Reduce is implemented as log(n) maps from lists of value pairs to lists of values.
// Each map corresponds to one level of a binary tree.
//
// The input x is a slice of type []T and the output y is a pointer of type *T.
// The function f is identified by name and must already be registered.
// The task should map a Pair of elements of type T to a single T.
//
// Reduce calls ReduceWriteTo with DefaultCmdOut, DefaultCmdErr.
func Reduce(f string, y, x, p interface{}) error {
	return ReduceWriteTo(f, y, x, p, DefaultCmdOut, DefaultCmdErr)
}

func ReduceWriteTo(f string, y, x, p interface{}, cmdout, cmderr io.Writer) error {
	out, err := reduce(f, x, p, cmdout, cmderr)
	if err != nil {
		return err
	}
	reflect.ValueOf(y).Elem().Set(reflect.ValueOf(out))
	return nil
}

func reduce(f string, x, p interface{}, cmdout, cmderr io.Writer) (interface{}, error) {
	// If there is only one element, return it.
	// Panics if the input list was empty.
	xval := reflect.ValueOf(x)
	if xval.Len() < 2 {
		return xval.Index(0).Interface(), nil
	}
	y, err := halve(f, x, p, cmdout, cmderr)
	if err != nil {
		return nil, err
	}
	return reduce(f, y, p, cmdout, cmderr)
}

// Maps n elements to ceil(n/2) elements.
// The input x must be a slice.
// Returns a slice of the same type.
func halve(f string, x, p interface{}, cmdout, cmderr io.Writer) (interface{}, error) {
	xval := reflect.ValueOf(x)
	n := reflect.ValueOf(x).Len()
	floor, ceil := n/2, (n+1)/2

	pairs := make([]Pair, floor)
	for i := range pairs {
		a := xval.Index(2 * i).Interface()
		b := xval.Index(2*i + 1).Interface()
		pairs[i] = Pair{a, b}
	}

	// Make a slice to assign the results to.
	// If n is even, then n/2 == (n+1)/2.
	// If n is odd, then this includes capacity for the last element.
	yptr := reflect.New(reflect.TypeOf(x)).Interface()
	y := reflect.ValueOf(yptr).Elem()
	y.Set(reflect.MakeSlice(reflect.TypeOf(x), floor, ceil))
	if err := MapWriteTo(f, yptr, pairs, p, cmdout, cmderr); err != nil {
		return nil, err
	}
	// If there were an odd number of elements,
	// then bring the last one forward.
	if n%2 != 0 {
		y = reflect.Append(y, xval.Index(n-1))
	}
	return y.Interface(), nil
}

// Pair describes a pair of values.
//
// Reduce tasks are executed as maps from pairs to single elements.
type Pair struct {
	A, B interface{}
}

// ReduceFunc creates a reduce task from a function.
// A reduce task maps a Pair of values to a single value.
//
// The function f must take either two or three arguments and
// have either one or two return values.
// The first two arguments and the first return value must have the same type.
// This type must be concrete.
// The second return value must be assignable to error.
//
// Examples:
//	var (
//		sum   = dstrfn.ReduceFunc(func(x, y float64) float64 { return x + y })
//		norm  = dstrfn.ReduceFunc(func(x, y float64) float64 { return math.Sqrt(x*x + y*y) }
//		pnorm = dstrfn.ReduceFunc(func(x, y, p float64) float64 {
// 			return math.Pow(math.Pow(x, p)+math.Pow(y, p), 1/p)
// 		})
//	)
//
// Reduce functions usually have the properties
//	1. f(x[0:n]) = f(f(x[0:m]), f(x[m:n]))
//	2. f(x[i]) = x[i]
// and therefore we define them in terms of their two-input case alone.
//
// ReduceFunc tasks should be invoked using Reduce() not Map().
func ReduceFunc(f interface{}) Task {
	fval := reflect.ValueOf(f)
	if fval.Kind() != reflect.Func {
		panic(fmt.Sprintf("not func: %v", fval.Kind()))
	}
	if n := fval.Type().NumIn(); n < 2 || n > 3 {
		panic(fmt.Sprintf("number of arguments: %d", n))
	}
	if n := fval.Type().NumOut(); n < 1 || n > 2 {
		panic(fmt.Sprintf("number of return values: %d", n))
	}
	return &reduceFuncTask{f}
}

// Task defined by a function.
type reduceFuncTask struct {
	F interface{}
}

// Returns a Pair containing the types of the first two arguments.
func (t *reduceFuncTask) NewInput() interface{} {
	f := reflect.ValueOf(t.F)
	a := reflect.New(f.Type().In(0)).Interface()
	b := reflect.New(f.Type().In(1)).Interface()
	return &Pair{a, b}
}

// Returns a new object of the type of the third argument.
// Returns nil if there is no second argument.
func (t *reduceFuncTask) NewConfig() interface{} {
	f := reflect.ValueOf(t.F)
	if f.Type().NumIn() < 3 {
		return nil
	}
	return reflect.New(f.Type().In(2)).Interface()
}

// Returns a new object of the type of the first return value.
func (t *reduceFuncTask) NewOutput() interface{} {
	f := reflect.ValueOf(t.F)
	return reflect.New(f.Type().Out(0)).Interface()
}

// If function only takes one argument then p is ignored.
func (t *reduceFuncTask) Func(x, p interface{}) (interface{}, error) {
	f := reflect.ValueOf(t.F)
	// Get two elements from x. Panics if x is not a Pair.
	ab := x.(Pair)
	in := []reflect.Value{
		reflect.ValueOf(ab.A).Elem(),
		reflect.ValueOf(ab.B).Elem(),
	}
	// Only use third argument if function accepts one.
	if f.Type().NumIn() > 2 {
		in = append(in, reflect.ValueOf(p))
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
