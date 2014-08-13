package main

import (
	"flag"
	"fmt"
	//"math"
	"os"

	"github.com/jvlmdr/go-pbs-pro/dstrfn"
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

	sqr := dstrfn.Func(func(x float64) float64 { return x * x })
	// Call function with one argument.
	dstrfn.Register("square", sqr)
	// Map operation with no extra arguments.
	dstrfn.RegisterMap("square-map", true, sqr)
	// Call function with multiple arguments.
	dstrfn.Register("add-three", dstrfn.Func(
		func(x, y, z float64) float64 { return x + y + z },
	))
	// Map operation with one extra argument.
	dstrfn.RegisterMap("add-const", true, dstrfn.ConfigFunc(
		func(x, y float64) float64 { return x + y },
	))
	//	// Reduce operation with no extra arguments.
	//	dstrfn.Register("add", false, dstrfn.ReduceFunc(
	//		func(x, y float64) float64 { return x + y },
	//	))
	//	// Reduce operation with one extra argument.
	//	dstrfn.Register("norm", false, dstrfn.ReduceFunc(
	//		func(x, y, p float64) float64 {
	//			return math.Pow(math.Pow(x, p)+math.Pow(y, p), 1/p)
	//		},
	//	))

	dstrfn.RegisterMap("vec-2-norm", true, dstrfn.Func(Norm))
	dstrfn.RegisterMap("vec-p-norm", true, dstrfn.ConfigFunc(NormP))
	//dstrfn.Register("vec-add", false, dstrfn.ReduceFunc(AddVec))

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

	// Square a number.
	var q float64
	if err := dstrfn.CallFunc("square", &q, 3); err != nil {
		fmt.Fprintln(os.Stderr, "call:", err)
		os.Exit(1)
	}
	fmt.Println(q)

	// Add three numbers.
	var s float64
	if err := dstrfn.CallFunc("add-three", &s, 3, 4, 5); err != nil {
		fmt.Fprintln(os.Stderr, "call:", err)
		os.Exit(1)
	}
	fmt.Println(s)

	// Square all numbers.
	var y []float64
	if err := dstrfn.MapFunc("square-map", &y, x); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
		os.Exit(1)
	}
	fmt.Println(y)

	// Subtract a constant from all numbers.
	var z []float64
	if err := dstrfn.MapFunc("add-const", &z, x, -(n + 1)); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
		os.Exit(1)
	}
	fmt.Println(z)

	//	// Compute sum of all numbers in a list.
	//	var sum float64
	//	if err := dstrfn.Reduce("add", &sum, x, nil); err != nil {
	//		fmt.Fprintln(os.Stderr, "reduce:", err)
	//		os.Exit(1)
	//	}
	//	fmt.Println("sum:", sum)

	//	// Compute 1.5-norm.
	//	// Demonstrates reduce function with a parameter.
	//	var norm float64
	//	if err := dstrfn.Reduce("norm", &norm, x, 1.5); err != nil {
	//		fmt.Fprintln(os.Stderr, "reduce:", err)
	//		os.Exit(1)
	//	}
	//	fmt.Println("1.5-norm:", norm)

	// Compute 2-norm of each vector.
	var norms2 []float64
	if err := dstrfn.MapFunc("vec-2-norm", &norms2, vecs); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
		os.Exit(1)
	}
	fmt.Println("norms2:", norms2)

	// Compute 1-norm of each vector.
	var norms1 []float64
	if err := dstrfn.MapFunc("vec-p-norm", &norms1, vecs, 1); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
		os.Exit(1)
	}
	fmt.Println("norms1:", norms1)

	//	// Compute sum of all vectors.
	//	var vecsum *Vec
	//	if err := dstrfn.Reduce("vec-add", &vecsum, vecs, nil); err != nil {
	//		fmt.Fprintln(os.Stderr, "map:", err)
	//		os.Exit(1)
	//	}
	//	fmt.Println("vecsum:", vecsum)
}
