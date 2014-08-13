package dstrfn

import (
	"flag"
	"fmt"
)

var (
	tasks    = make(map[string]*taskSpec)
	mapTasks = make(map[string]*mapTaskSpec)
)

// Task for submission.
// Has a number of extra options.
type taskSpec struct {
	Task ConfigTask
	// Additional flags for the job.
	Flags string
	// Keep stdout and stderr of tasks?
	Stdout, Stderr bool
}

type mapTaskSpec struct {
	taskSpec
	// Group jobs into chunks?
	// Chunk is set in Register(), ChunkLen is set by a flag.
	Chunk    bool
	ChunkLen int
}

// Registers a task to a name.
// The name must be able to be part of a command-line flag.
// The task must implement Task or ConfigTask.
func Register(name string, task interface{}) {
	register(name, toConfigTask(task))
}

// Chunking is only supported for "simple" types.
// That is, types X which can be decoded from JSON into new([]X).
func RegisterMap(name string, chunk bool, task interface{}) {
	registerMap(name, chunk, toConfigTask(task))
}

func register(name string, task ConfigTask) {
	if nameUsed(name) {
		panic(fmt.Sprintf(`name already registered: "%s"`, name))
	}
	spec := &taskSpec{Task: task}
	registerSpecFlags(name, spec)
	tasks[name] = spec
}

func registerMap(name string, chunk bool, task ConfigTask) {
	if nameUsed(name) {
		panic(fmt.Sprintf(`name already registered: "%s"`, name))
	}
	if chunk {
		task = &chunkTask{task}
	}
	spec := new(mapTaskSpec)
	spec.Task = task
	spec.Chunk = chunk
	registerSpecFlags(name, &spec.taskSpec)
	flag.IntVar(&spec.ChunkLen, name+".chunk-len", 1, "Split into chunks of up to this many elements.")
	mapTasks[name] = spec
}

func nameUsed(name string) bool {
	if _, used := tasks[name]; used {
		return true
	}
	if _, used := mapTasks[name]; used {
		return true
	}
	return false
}

func registerSpecFlags(name string, spec *taskSpec) {
	flag.StringVar(&spec.Flags, name+".flags", "", "Additional flags")
	flag.BoolVar(&spec.Stdout, name+".stdout", false, "Keep stdout?")
	flag.BoolVar(&spec.Stderr, name+".stderr", false, "Keep stderr?")
}

func toConfigTask(task interface{}) ConfigTask {
	switch task := task.(type) {
	case Task:
		return configTask{task}
	case ConfigTask:
		return task
	default:
		panic("task does not implement Task or ConfigTask")
	}
}
