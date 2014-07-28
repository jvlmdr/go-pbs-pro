package dstrfn

import (
	"flag"
	"fmt"
)

var (
	tasks     map[string]*qsubTask
	addrStr   string
	slaveTask string
)

func init() {
	tasks = make(map[string]*qsubTask)
	flag.StringVar(&addrStr, "dstrfn.addr", "", "Address of master on network.")
	flag.StringVar(&slaveTask, "dstrfn.task", "", "Task to execute as slave. Empty to execute as master.")
}

type qsubTask struct {
	Task Task
	// Additional flags for the job.
	Flags string
	// Group jobs into chunks.
	ChunkLen int
	// Keep stdout and stderr of tasks?
	Stdout, Stderr bool
}

// Registers a task to a name.
// The name must be able to be part of a command-line flag.
func Register(name string, task Task) {
	_, used := tasks[name]
	if used {
		panic(fmt.Sprintf(`name already registered: "%s"`, name))
	}

	q := new(qsubTask)
	//q.Task = &chunkTask{task}
	q.Task = task
	flag.StringVar(&q.Flags, name+".flags", "", "Additional flags")
	//flag.IntVar(&q.ChunkLen, name+".chunk-len", 1, "Split into chunks of up to this many elements")
	flag.BoolVar(&q.Stdout, name+".stdout", false, "Keep stdout?")
	flag.BoolVar(&q.Stderr, name+".stderr", false, "Keep stderr?")
	tasks[name] = q
}
