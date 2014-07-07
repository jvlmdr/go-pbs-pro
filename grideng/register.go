package grideng

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
	flag.StringVar(&addrStr, "grideng.addr", "", "Address of master on network.")
	flag.StringVar(&slaveTask, "grideng.task", "", "Task to execute as slave. Empty to execute as master.")
}

type qsubTask struct {
	Task Task
	// Resources string (-l) for qsub.
	Res string
	// Group jobs into chunks.
	ChunkLen int
	// Where to route stdout and stderr of tasks.
	Stdout, Stderr string
}

// Registers a task to a name.
// The name must be able to be part of a command-line flag.
func Register(name string, task Task) {
	_, used := tasks[name]
	if used {
		panic(fmt.Sprintf(`name already registered: "%s"`, name))
	}

	q := new(qsubTask)
	q.Task = &chunkTask{task}
	flag.StringVar(&q.Res, name+".l", "", "Resource flag (-l) to qsub")
	flag.IntVar(&q.ChunkLen, name+".chunk-len", 1, "Split into chunks of up to this many elements")
	flag.StringVar(&q.Stdout, name+".stdout", "/dev/null", "Where to save stdout of task")
	flag.StringVar(&q.Stderr, name+".stderr", "/dev/null", "Where to save stderr of task")
	tasks[name] = q
}
