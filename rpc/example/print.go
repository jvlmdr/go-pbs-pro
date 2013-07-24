package main

import (
	"fmt"
	"github.com/jackvalmadre/go-grideng/rpc"
)

type PrintMap struct {
	X []float64
}

func (m PrintMap) Len() int                 { return len(m.X) }
func (m PrintMap) Input(i int) interface{}  { return m.X[i] }
func (m PrintMap) Output(i int) interface{} { return new(struct{}) }

func (m PrintMap) Task(i int) grideng.Task {
	return &PrintTask{m.X[i]}
}

type PrintTask struct{ X float64 }

func (t *PrintTask) Input() interface{} { return &t.X }
func (t PrintTask) Output() interface{} { return struct{}{} }

func (t *PrintTask) Do() error {
	// This is the actual work.
	fmt.Println("The number is...", t.X)
	return nil
}
