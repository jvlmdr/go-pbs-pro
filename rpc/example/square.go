package main

import "github.com/jackvalmadre/go-grideng/rpc"

type SquareMap struct {
	X []float64
}

func (m SquareMap) Len() int                 { return len(m.X) }
func (m SquareMap) Input(i int) interface{}  { return m.X[i] }
func (m SquareMap) Output(i int) interface{} { return &m.X[i] }

func (m SquareMap) Task(i int) grideng.Task {
	return &SquareTask{m.X[i]}
}

type SquareTask struct{ X float64 }

func (t *SquareTask) Input() interface{} { return &t.X }
func (t SquareTask) Output() interface{} { return t.X }

func (t *SquareTask) Do() error {
	// This is the actual work.
	t.X = t.X * t.X
	return nil
}
