package grideng

import "fmt"

// Computes y[i] = f(x[i], p) for all i.
//
// Parameters:
// 	name identifies the registered task f.
// 	y is a slice to which the outputs will be assigned.
//	x is a slice of inputs.
// 	p contains optional constant parameters to f.
func Map(name string, y, x, p interface{}) error {
	task, there := tasks[name]
	if !there {
		return fmt.Errorf(`task not found: "%s"`, name)
	}
	return master(task, name, y, x, p)
}
