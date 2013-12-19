package grideng

import (
	"fmt"
	"reflect"
)

// Reduce is implemented as log(n) maps from lists of pairs to single values.
// Each map corresponds to one level of a binary tree.
//
// x is a slice of type []T.
// name corresponds to a task which takes a pair of objects of type T
// and returns one object of type T.
// y is a reference to a single output value (e.g. pointer, interface).
// p is an optional configuration parameter.
func Reduce(name string, y, x, p interface{}) error {
	out, err := reduce(name, x, p)
	if err != nil {
		return err
	}
	reflect.ValueOf(y).Elem().Set(reflect.ValueOf(out))
	return nil
}

func reduce(name string, x, p interface{}) (interface{}, error) {
	// If there is only one element, return it.
	// Panics if the input list was empty.
	xval := reflect.ValueOf(x)
	if xval.Len() < 2 {
		return xval.Index(0).Interface(), nil
	}
	y, err := halve(name, x, p)
	if err != nil {
		return nil, err
	}
	return reduce(name, y, p)
}

// Maps n elements to ceil(n/2) elements.
func halve(name string, x, p interface{}) (interface{}, error) {
	xval := reflect.ValueOf(x)
	n := reflect.ValueOf(x).Len()
	floor, ceil := n/2, (n+1)/2

	pairs := make([]*pair, floor)
	for i := range pairs {
		a := xval.Index(2 * i).Interface()
		b := xval.Index(2*i + 1).Interface()
		pairs[i] = &pair{a, b}
	}

	// Make a slice to assign the results to.
	// If n is even, then n/2 == (n+1)/2.
	// If n is odd, then this includes capacity for the last element.
	y := make([]interface{}, len(pairs), ceil)
	if err := Map(name, y, pairs, p); err != nil {
		return nil, err
	}
	// If there were an odd number of elements,
	// then bring the last one forward.
	if n%2 != 0 {
		y = append(y, xval.Index(n-1).Interface())
	}
	return y, nil
}

// Reduce is performed as a series of maps on lists of pairs.
type pair struct {
	A, B interface{}
}

// Defines a reduce task from a function.
// A reduce task maps pairs of values to a single value.
//
// The function f must take either two or three arguments and
// have either one or two return values.
// The first two arguments and the first return value must have the same type.
// This type must be concrete.
// The second return value must be assignable to error.
//
// Examples:
//	sum := grideng.ReduceFunc(func(x, y float64) float64 { return x + y })
//	norm := grideng.ReduceFunc(func(x, y float64) float64 { return math.Sqrt(x*x + y*y) }
//	pnorm := grideng.ReduceFunc(func(x, y, p float64) float64 {
// 		return math.Pow(math.Pow(x, p)+math.Pow(y, p), 1/p)
// 	})
//
// Reduce functions usually have the properties
//	1. f(x[0:n]) = f(f(x[0:m]), f(x[m:n]))
//	             = f(...f(f(x[0:2]), f(x[2:4])), ...)
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

// Returns a pair containing the types of the first two arguments.
func (t *reduceFuncTask) NewInput() interface{} {
	f := reflect.ValueOf(t.F)
	a := reflect.New(f.Type().In(0)).Interface()
	b := reflect.New(f.Type().In(1)).Interface()
	return &pair{a, b}
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
	// Get two elements from x. Panics if x is not a *pair.
	ab := x.(*pair)
	in := []reflect.Value{
		reflect.ValueOf(ab.A).Elem(),
		reflect.ValueOf(ab.B).Elem(),
	}
	// Only use third argument if function accepts one.
	if f.Type().NumIn() > 2 {
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
