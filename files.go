package grideng

import "fmt"

// Returns the output file using the task's number.
// Supplied with the Grid Engine task ID (one-indexed not zero-indexed).
func inputFile(num int) string {
	return fmt.Sprintf("in-%010d", num)
}

// Returns the output file using the task's unique name.
func outputFile(name string) string {
	return fmt.Sprintf("out-%s", name)
}
