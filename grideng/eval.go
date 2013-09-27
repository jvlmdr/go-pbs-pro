package grideng

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

// Evaluates y, err := f(x).
// y must be a pointer.
func Eval(y, f, x interface{}) error {
	// Construct argument list.
	in := []reflect.Value{reflect.ValueOf(x)}
	// Panics if f is not a function or inputs are invalid.
	out := reflect.ValueOf(f).Call(in)
	// Second return value must be nil.
	// Panics if there was less than two values or value was not an error.
	if err := nonNilErr(out[1].Interface()); err != nil {
		return err
	}
	// Panics if y is not a pointer or return value is not assignable to *y.
	reflect.ValueOf(y).Elem().Set(out[0])
	return nil
}

// If x is non-nil, converts to an error and returns it.
// If x is nil, returns nil.
func nonNilErr(x interface{}) error {
	if x == nil {
		return nil
	}
	err := x.(error)
	return err
}

// Decodes x from in, evaluates y, err := f(x), and encodes the result to out.
func ReadEvalWrite(out io.Writer, f interface{}, in io.Reader) error {
	// Read input.
	x := reflect.New(reflect.TypeOf(f).In(0))
	if err := json.NewDecoder(in).Decode(x.Interface()); err != nil {
		return errors.New("decode input: " + err.Error())
	}

	// Perform function.
	y := reflect.New(reflect.TypeOf(f).Out(0))
	if err := Eval(y.Interface(), f, x.Elem().Interface()); err != nil {
		return err
	}

	// Write output.
	if err := json.NewEncoder(out).Encode(y.Interface()); err != nil {
		return errors.New("encode output: " + err.Error())
	}
	return nil
}
