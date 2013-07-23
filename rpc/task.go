package main

// May operate in-place, therefore only Input() should be used before Do() and only Output() should be used after Do().
type Task interface {
	// Must contain a non-nil pointer for use with Decode.
	Input() interface{}
	Do() error
	// Can contain a pointer or a value, doesn't matter for Decode.
	Output() interface{}
}
