package dstrfn

import (
	"fmt"
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

// Assumes that len(dst) = sum_i len(src[i]).
func mergeTo(dst, src interface{}) {
	srcval := reflect.ValueOf(src)
	dstval := reflect.ValueOf(dst)
	m := srcval.Len()
	// Count number of elements.
	var n int
	for i := 0; i < m; i++ {
		n += srcval.Index(i).Len()
	}
	if dstval.Len() != n {
		panic(fmt.Sprintf("slice len: expect %d, got %d", n, dstval.Len()))
	}
	if m == 0 {
		return
	}
	// Copy elements.
	p := srcval.Index(0).Len()
	ind := 0
	for j := 0; j < p; j++ {
		for i := 0; i < m; i++ {
			xi := srcval.Index(i)
			if j >= xi.Len() {
				break
			}
			dstval.Index(ind).Set(xi.Index(j))
			ind++
		}
	}
}
