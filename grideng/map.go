package grideng

type Map interface {
	Len() int
	// Returns the i-th input.
	Input(i int) interface{}
	// Returns a pointer to the i-th output, suitable for use with Decode.
	Output(i int) interface{}
}
