package dstrfn

import (
	"reflect"
)

// Takes an input of []X and returns an input of [][]X.
func split(x interface{}, minNum, maxSize int) interface{} {
	xval := reflect.ValueOf(x)
	n := xval.Len()
	// Split m into the largest groups allowed
	// but do not allow there to be too few groups.
	// Number of groups cannot exceed number of elements.
	m := max(ceilDiv(n, maxSize), min(minNum, n))
	y := reflect.MakeSlice(reflect.SliceOf(xval.Type()), m, m)
	for i := 0; i < m; i++ {
		yi := reflect.MakeSlice(xval.Type(), 0, ceilDiv(n, m))
		for j := 0; m*j+i < n; j++ {
			yi = reflect.Append(yi, xval.Index(m*j+i))
		}
		y.Index(i).Set(yi)
	}
	return y.Interface()
}

// Takes a slice [][]X and returns a slice []X.
func merge(x interface{}) interface{} {
	xval := reflect.ValueOf(x)
	m := xval.Len()
	if m == 0 {
		return nil
	}

	p := xval.Index(0).Len()
	y := reflect.MakeSlice(xval.Type().Elem(), 0, m*p)
	for j := 0; j < p; j++ {
		for i := 0; i < m; i++ {
			xi := xval.Index(i)
			if j >= xi.Len() {
				break
			}
			y = reflect.Append(y, xi.Index(j))
		}
	}
	return y.Interface()
}

// Task defined by a function.
type chunkTask struct {
	ElemTask Task
}

// Creates a new element input, discards it,
// and creates a new slice of the type that it pointed to.
func (t *chunkTask) NewInput() interface{} {
	e := t.ElemTask.NewInput()
	etyp := reflect.TypeOf(e).Elem()
	return reflect.New(reflect.SliceOf(etyp)).Interface()
}

// Returns a new object of the type of the second argument.
// Returns nil if there is no second argument.
func (t *chunkTask) NewConfig() interface{} {
	return t.ElemTask.NewConfig()
}

// Returns a new object of the type of the first return value.
func (t *chunkTask) NewOutput() interface{} {
	e := t.ElemTask.NewOutput()
	etyp := reflect.TypeOf(e).Elem()
	return reflect.New(reflect.SliceOf(etyp)).Interface()
}

// If function only takes one argument then p is ignored.
func (t *chunkTask) Func(x, p interface{}) (interface{}, error) {
	xval := reflect.ValueOf(x)
	n := xval.Len()
	ytyp := reflect.TypeOf(t.NewOutput()).Elem()
	y := reflect.MakeSlice(ytyp, n, n)

	for i := 0; i < n; i++ {
		xi := xval.Index(i).Interface()
		yi, err := t.ElemTask.Func(xi, p)
		if err != nil {
			return nil, err
		}
		y.Index(i).Set(reflect.ValueOf(yi))
	}
	return y.Interface(), nil
}
