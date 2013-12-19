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
	Res *string
}

// Registers a task to a name.
// The name must be able to be part of a command-line flag.
func Register(name string, task Task) {
	_, used := tasks[name]
	if used {
		panic(fmt.Sprintf(`name already registered: "%s"`, name))
	}

	q := new(qsubTask)
	q.Task = task
	q.Res = flag.String(name+".l", "", "Resource flag (-l) to qsub")
	tasks[name] = q
}
