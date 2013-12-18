package grideng

import "fmt"

// Computes y[i] = f(x[i], p) for all i.
//
// x is a slice of inputs.
// y is a slice to which the outputs will be assigned.
// f is a registered task identified by name.
// p contains (optional) constant parameters to f.
func Map(name string, y, x, p interface{}) error {
	task, there := tasks[name]
	if !there {
		panic(fmt.Sprintf(`task not found: "%s"`, name))
	}
	return master(task, name, y, x, p)
}
