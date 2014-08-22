package dstrfn

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"

	"github.com/jvlmdr/go-file/fileutil"
)

var (
	workerTask   string
	workerDir    string
	workerMapLen int
)

func init() {
	flag.StringVar(&workerTask, "dstrfn.task", "", "Task to execute as slave. Empty to execute as master.")
	flag.StringVar(&workerDir, "dstrfn.dir", "", "Location of temporary files.")
	flag.IntVar(&workerMapLen, "dstrfn.map", 0, "The number of tasks in the map. Zero if not a map operation.")
}

// If the process is a worker, this function never returns.
func ExecIfSlave() {
	if len(workerTask) == 0 {
		// Not a worker.
		return
	}
	if err := worker(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func worker() error {
	// Change current directory to that of submission.
	wd, err := getenv("PBS_O_WORKDIR")
	if err != nil {
		return err
	}
	if err := os.Chdir(wd); err != nil {
		return fmt.Errorf("chdir: %v", err)
	}

	// Determine file locations.
	var inFile, outFile, errFile string
	if workerMapLen > 0 {
		// If this is a map task, then use the array index.
		var ind int
		// Array index cannot be set for maps of 1 job.
		// In this case the index is zero.
		if workerMapLen > 1 {
			var err error
			ind, err = getenvInt("PBS_ARRAY_INDEX")
			if err != nil {
				return err
			}
			// Convert to zero-indexed.
			ind--
		}
		inFile = fmt.Sprintf("in-%d.json", ind)
		outFile = fmt.Sprintf("out-%d.json", ind)
		errFile = fmt.Sprintf("err-%d.json", ind)
	} else {
		inFile = "in.json"
		outFile = "out.json"
		errFile = "err.json"
	}
	inFile = path.Join(workerDir, inFile)
	outFile = path.Join(workerDir, outFile)
	errFile = path.Join(workerDir, errFile)
	// Config file does not vary with index.
	confFile := path.Join(workerDir, "conf.json")

	// Error can only be communicated once the task ID has been determined.
	if err := doTask(inFile, confFile, outFile); err != nil {
		// Attempt to save error.
		if err := fileutil.SaveExt(errFile, err.Error()); err != nil {
			return err
		}
	}
	return nil
}

// An error returned by this function will be communicated to the master.
// Or at least we will try.
// This can only be done once the task ID has been determined.
func doTask(inFile, confFile, outFile string) error {
	// Look up task by name.
	var task ConfigTask
	if workerMapLen > 0 {
		spec, there := mapTasks[workerTask]
		if !there {
			return fmt.Errorf(`map task not found: "%s"`, workerTask)
		}
		task = spec.Task
	} else {
		spec, there := tasks[workerTask]
		if !there {
			return fmt.Errorf(`task not found: "%s"`, workerTask)
		}
		task = spec.Task
	}

	x := task.NewInput()
	if x != nil {
		log.Println("load input:", inFile)
		if err := fileutil.LoadExt(inFile, x); err != nil {
			return err
		}
		x = deref(x)
	}
	p := task.NewConfig()
	if p != nil {
		log.Println("load config:", confFile)
		if err := fileutil.LoadExt(confFile, p); err != nil {
			return err
		}
		p = deref(p)
	}
	log.Println("call function")
	y, err := task.Func(x, p)
	if err != nil {
		return err
	}
	if y != nil {
		log.Println("save output:", outFile)
		if err := fileutil.SaveExt(outFile, y); err != nil {
			return err
		}
	}
	return nil
}

func getenv(name string) (string, error) {
	val := os.Getenv(name)
	if len(val) == 0 {
		return "", fmt.Errorf("environment variable empty: %s", name)
	}
	return val, nil
}

func getenvInt(name string) (int, error) {
	str, err := getenv(name)
	if err != nil {
		return 0, err
	}
	x, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(x), nil
}
