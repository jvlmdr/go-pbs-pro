package main

import (
	"github.com/jackvalmadre/go-grideng/grideng"

	"flag"
	"fmt"
	"math"
	"os"
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

	grideng.Register("square", grideng.Func(func(x float64) float64 { return x * x }))
	grideng.Register("add-const", grideng.Func(func(x, y float64) float64 { return x + y }))
	grideng.Register("add", grideng.ReduceFunc(func(x, y float64) float64 { return x + y }))
	grideng.Register("norm", grideng.ReduceFunc(func(x, y, p float64) float64 {
		return math.Pow(math.Pow(x, p)+math.Pow(y, p), 1/p)
	}))

	grideng.Register("vec-2-norm", grideng.Func(Norm))
	grideng.Register("vec-p-norm", grideng.Func(NormP))
	grideng.Register("add-vec", grideng.ReduceFunc(AddVec))

	flag.Parse()
	grideng.ExecIfSlave()

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
	if err := grideng.Map("square", y, x, nil); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println(y)

	// Subtract a constant from all numbers.
	z := make([]float64, n)
	if err := grideng.Map("add-const", z, x, -(n + 1)); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println(z)

	// Compute sum of all numbers in a list.
	var sum float64
	if err := grideng.Reduce("add", &sum, x, nil); err != nil {
		fmt.Fprintln(os.Stderr, "reduce:", err)
	}
	fmt.Println("sum:", sum)

	// Compute 1.5-norm.
	// Demonstrates reduce function with a parameter.
	var norm float64
	if err := grideng.Reduce("norm", &norm, x, 1.5); err != nil {
		fmt.Fprintln(os.Stderr, "reduce:", err)
	}
	fmt.Println("1.5-norm:", norm)

	// Compute 2-norm of each vector.
	norms2 := make([]float64, m)
	if err := grideng.Map("vec-2-norm", norms2, vecs, nil); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println("norms2:", norms2)

	// Compute 1-norm of each vector.
	norms1 := make([]float64, m)
	if err := grideng.Map("vec-p-norm", norms1, vecs, 1); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println("norms1:", norms1)

	// Compute sum of all vectors.
	var vecsum *Vec
	if err := grideng.Reduce("add-vec", &vecsum, vecs, nil); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println("vecsum:", sum)
}
