package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/jvlmdr/go-drmaa-util/dstrfn"
)

func main() {
	var (
		n int
		m int
		d int
	)
	flag.IntVar(&n, "n", 8, "Sum squares from 1 to n")
	flag.IntVar(&m, "m", 8, "Number of vectors")
	flag.IntVar(&d, "d", 8, "Number of dimensions for vector")

	dstrfn.Register("square", dstrfn.Func(
		func(x float64) float64 { return x * x },
	))
	dstrfn.Register("add-const", dstrfn.Func(
		func(x, y float64) float64 { return x + y },
	))
	dstrfn.Register("add", dstrfn.ReduceFunc(
		func(x, y float64) float64 { return x + y },
	))
	dstrfn.Register("norm", dstrfn.ReduceFunc(
		func(x, y, p float64) float64 {
			return math.Pow(math.Pow(x, p)+math.Pow(y, p), 1/p)
		},
	))

	dstrfn.Register("vec-2-norm", dstrfn.Func(Norm))
	dstrfn.Register("vec-p-norm", dstrfn.Func(NormP))
	dstrfn.Register("vec-add", dstrfn.ReduceFunc(AddVec))

	flag.Parse()
	dstrfn.ExecIfSlave()

	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = float64(i + 1)
	}

	vecs := make([]*Vec, m)
	for i := range vecs {
		vecs[i] = RandVec(d)
	}

	// Square all numbers.
	y := make([]float64, n)
	if err := dstrfn.Map("square", y, x, nil); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println(y)

	// Subtract a constant from all numbers.
	z := make([]float64, n)
	if err := dstrfn.Map("add-const", z, x, -(n + 1)); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println(z)

	// Compute sum of all numbers in a list.
	var sum float64
	if err := dstrfn.Reduce("add", &sum, x, nil); err != nil {
		fmt.Fprintln(os.Stderr, "reduce:", err)
	}
	fmt.Println("sum:", sum)

	// Compute 1.5-norm.
	// Demonstrates reduce function with a parameter.
	var norm float64
	if err := dstrfn.Reduce("norm", &norm, x, 1.5); err != nil {
		fmt.Fprintln(os.Stderr, "reduce:", err)
	}
	fmt.Println("1.5-norm:", norm)

	// Compute 2-norm of each vector.
	norms2 := make([]float64, m)
	if err := dstrfn.Map("vec-2-norm", norms2, vecs, nil); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println("norms2:", norms2)

	// Compute 1-norm of each vector.
	norms1 := make([]float64, m)
	if err := dstrfn.Map("vec-p-norm", norms1, vecs, 1); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println("norms1:", norms1)

	// Compute sum of all vectors.
	var vecsum *Vec
	if err := dstrfn.Reduce("vec-add", &vecsum, vecs, nil); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println("vecsum:", vecsum)
}
