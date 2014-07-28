package dstrfn

// Describes a task to execute on a remote host.
//
// Captures the minimum required by master and slave.
//
// NewInput(), NewConfig() and Func() are called by the slave,
// NewOutput() is called by the master.
//
// NewX() must return a value suitable for decoding into.
type Task interface {
	// Returns an input object which can be decoded into.
	//	x := task.NewInput()
	//	json.NewDecoder(r).Decode(x)
	// This will be de-referenced before passing to Func().
	NewInput() interface{}
	// Returns a config object which can be decoded into,
	// or nil if there are no config parameters.
	NewConfig() interface{}

	// Performs the task given input x and parameters p.
	Func(x, p interface{}) (interface{}, error)

	// Returns an output object which can be decoded into.
	//	y := task.NewOutput()
	//	json.NewDecoder(r).Decode(y)
	NewOutput() interface{}
}
