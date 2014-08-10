package main

import (
	"math"
	"math/rand"
)

type Vec struct {
	Elems []float64
}

func NewVec(n int) *Vec {
	return &Vec{make([]float64, n)}
}

func Norm(x *Vec) float64 {
	return NormP(x, 2)
}

func NormP(x *Vec, p float64) float64 {
	var r float64
	for i := range x.Elems {
		r += math.Pow(math.Abs(x.Elems[i]), p)
	}
	return math.Pow(r, 1/p)
}

func RandVec(n int) *Vec {
	x := NewVec(n)
	for i := range x.Elems {
		x.Elems[i] = rand.NormFloat64()
	}
	return x
}

func AddVec(x, y *Vec) *Vec {
	z := NewVec(len(x.Elems))
	for i := range z.Elems {
		z.Elems[i] = x.Elems[i] + y.Elems[i]
	}
	return z
}
