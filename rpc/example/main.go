package main

import (
	"flag"
	"fmt"
	"github.com/jackvalmadre/go-grideng/rpc"
	"os"
)

type SquareMap struct {
	X []float64
}

func (m SquareMap) Len() int                 { return len(m.X) }
func (m SquareMap) Input(i int) interface{}  { return m.X[i] }
func (m SquareMap) Output(i int) interface{} { return &m.X[i] }

func (m SquareMap) Task(i int) rpc.Task {
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

func main() {
	// Program flags.
	var n int
	flag.IntVar(&n, "n", 0, "Number of jobs")
	// Grid engine flags.
	var (
		master    bool
		slave     bool
		port      string
		addr      string
		codec     string
		resources string
	)
	flag.BoolVar(&master, "master", false, "Operate in master mode?")
	flag.BoolVar(&slave, "slave", false, "Operate in slave mode?")
	flag.StringVar(&addr, "addr", "", "Master address")
	flag.StringVar(&port, "port", "1234", "Master port")
	flag.StringVar(&codec, "codec", "json", "Codec (json or gob)")
	flag.StringVar(&resources, "l", "", "Grid engine resources (qsub -l flag)")

	flag.Parse()

	if slave && !master {
		var task SquareTask
		rpc.ExecSlave(&task, addr, port, codec)
	}

	if slave == master {
		fmt.Fprintln(os.Stderr, "Must specify master xor slave")
		os.Exit(1)
	}

	if n <= 0 {
		fmt.Fprintln(os.Stderr, "Require n > 0")
		os.Exit(1)
	}
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = float64(i)
	}
	m := SquareMap{x}

	rpc.Do(m, addr, port, codec, resources)
	fmt.Println(x)
}
