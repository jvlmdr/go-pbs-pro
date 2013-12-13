package main

import (
	"flag"
	"fmt"
	"github.com/jackvalmadre/go-grideng/rpc"
	"log"
	"os"
)

func main() {
	// Program flags.
	var n int
	flag.IntVar(&n, "n", 0, "Number of jobs")
	// Grid engine flags.
	var (
		master    bool
		slave     bool
		mode      string
		addr      string
		codec     string
		resources string
	)
	flag.BoolVar(&master, "master", false, "Operate in master mode?")
	flag.BoolVar(&slave, "slave", false, "Operate in slave mode?")
	flag.StringVar(&mode, "mode", "", `"square" or "print"`)
	flag.StringVar(&addr, "addr", "", "Address of server")
	flag.StringVar(&codec, "codec", "json", "Codec (json or gob)")
	flag.StringVar(&resources, "l", "", "Grid engine resources (qsub -l flag)")

	flag.Parse()

	if slave && !master {
		var task grideng.Task
		switch mode {
		case "square":
			task = new(SquareTask)
		case "print":
			task = new(PrintTask)
		default:
			fmt.Fprintf(os.Stderr, "Invalid mode \"%s\"\n", mode)
			os.Exit(1)
		}
		grideng.ExecSlave(task, addr, codec)
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

	var err error

	square := SquareMap{x}
	err = grideng.Do(square, addr, codec, resources, []string{"-slave", "-mode=square"})
	if err != nil {
		log.Fatal(err)
	}

	prnt := PrintMap{x}
	err = grideng.Do(prnt, addr, codec, resources, []string{"-slave", "-mode=print"})
	if err != nil {
		log.Fatal(err)
	}
}
