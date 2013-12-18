package main

import (
	"github.com/jackvalmadre/go-grideng/grideng"

	"flag"
	"fmt"
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

	grideng.Register("square", &grideng.Func{func(x float64) float64 { return x * x }})
	grideng.Register("add-const", &grideng.Func{func(x, y float64) float64 { return x + y }})
	grideng.Register("add", &grideng.ReduceFunc{func(x, y float64) float64 { return x + y }})

	grideng.Register("vec-2-norm", &grideng.Func{Norm})
	grideng.Register("vec-p-norm", &grideng.Func{NormP})

	flag.Parse()
	grideng.ExecIfSlave()

	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = float64(i + 1)
	}

	// y[i] <- x[i]^2
	y := make([]float64, n)
	if err := grideng.Map("square", y, x, nil); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println(y)

	// z[i] <- x[i] - (n+1)
	z := make([]float64, n)
	if err := grideng.Map("add-const", z, x, -(n + 1)); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println(z)

	// total <- sum_{i} x[i]
	var total float64
	if err := grideng.Reduce("add", &total, x, nil); err != nil {
		fmt.Fprintln(os.Stderr, "reduce:", err)
	}
	fmt.Println(total)

	vecs := make([]*Vec, m)
	for i := range vecs {
		vecs[i] = RandVec(d)
	}

	// norms2[i] <- Norm(vecs[i])
	norms2 := make([]float64, m)
	if err := grideng.Map("vec-2-norm", norms2, vecs, nil); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println("norms2:", norms2)

	// norms1[i] <- NormP(vecs[i], 1)
	norms1 := make([]float64, m)
	if err := grideng.Map("vec-p-norm", norms1, vecs, 1); err != nil {
		fmt.Fprintln(os.Stderr, "map:", err)
	}
	fmt.Println("norms1:", norms1)
}
