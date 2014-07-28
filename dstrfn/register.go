package dstrfn

import (
	"flag"
	"fmt"
)

var (
	tasks     map[string]*subTask
	addrStr   string
	slaveTask string
)

func init() {
	tasks = make(map[string]*subTask)
	flag.StringVar(&addrStr, "dstrfn.addr", "", "Address of master on network.")
	flag.StringVar(&slaveTask, "dstrfn.task", "", "Task to execute as slave. Empty to execute as master.")
}

// Task for submission.
// Has a number of extra options.
type subTask struct {
	Task Task
	// Additional flags for the job.
	Flags string
	// Group jobs into chunks.
	Chunk    bool
	ChunkLen int
	// Keep stdout and stderr of tasks?
	Stdout, Stderr bool
}

// Registers a task to a name.
// The name must be able to be part of a command-line flag.
//
// Chunking is only supported for "simple" types.
// That is, types X which can be decoded from JSON into new([]X).
func Register(name string, chunk bool, task Task) {
	_, used := tasks[name]
	if used {
		panic(fmt.Sprintf(`name already registered: "%s"`, name))
	}

	st := new(subTask)
	if chunk {
		st.Task = &chunkTask{task}
		st.Chunk = true
		flag.IntVar(&st.ChunkLen, name+".chunk-len", 1, "Split into chunks of up to this many elements")
	} else {
		st.Task = task
	}
	flag.StringVar(&st.Flags, name+".flags", "", "Additional flags")
	flag.BoolVar(&st.Stdout, name+".stdout", false, "Keep stdout?")
	flag.BoolVar(&st.Stderr, name+".stderr", false, "Keep stderr?")
	tasks[name] = st
}
