package dstrfn

import (
	"log"
	"reflect"
)

// A meta-task which performs another task on multiple elements.
type chunkTask struct {
	Task ConfigTask
}

// Creates a new input element, discards it,
// and creates a new slice of the type that it pointed to.
func (t *chunkTask) NewInput() interface{} {
	e := t.Task.NewInput()
	etyp := reflect.TypeOf(e).Elem()
	return reflect.New(reflect.SliceOf(etyp)).Interface()
}

// Returns a new object of the type of the second argument.
// Returns nil if there is no second argument.
func (t *chunkTask) NewConfig() interface{} {
	return t.Task.NewConfig()
}

// Returns a new object of the type of the first return value.
func (t *chunkTask) NewOutput() interface{} {
	ptr := reflect.ValueOf(t.Task.NewOutput())
	if ptr.Type().Kind() != reflect.Ptr {
		log.Println("not a pointer:", ptr.Type())
	}
	xtyp := ptr.Type().Elem()
	return reflect.New(reflect.SliceOf(xtyp)).Interface()
}

// If function only takes one argument then p is ignored.
func (t *chunkTask) Func(x, p interface{}) (interface{}, error) {
	xval := reflect.ValueOf(x)
	n := xval.Len()
	ytyp := reflect.TypeOf(t.NewOutput()).Elem()
	y := reflect.MakeSlice(ytyp, n, n)

	for i := 0; i < n; i++ {
		xi := xval.Index(i).Interface()
		yi, err := t.Task.Func(xi, p)
		if err != nil {
			return nil, err
		}
		y.Index(i).Set(reflect.ValueOf(yi))
	}
	return y.Interface(), nil
}
