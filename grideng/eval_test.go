package grideng

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"testing"
)

var sqr = func(x float64) (float64, error) {
	return x * x, nil
}

var sqrt = func(x float64) (float64, error) {
	if x < 0 {
		return 0, errors.New("negative input")
	}
	return math.Sqrt(x), nil
}

// Evaluate a simple function.
func TestEval(t *testing.T) {
	x := float64(5)
	var y float64
	err := Eval(&y, sqr, x)
	if err != nil {
		t.Fatal(err)
	}
	checkEqual(t, 25, y)
}

// Evaluate a function which encounters an error.
func TestEval_error(t *testing.T) {
	x := float64(-1)
	var y float64
	err := Eval(&y, sqrt, x)
	if err == nil {
		t.Fatal("expected to receive error")
	}
}

// Evaluate a simple function.
// Pass arugments and receive result by JSON.
func TestReadEvalWrite(t *testing.T) {
	x := float64(5)
	// Write input to stream.
	var in bytes.Buffer
	if err := json.NewEncoder(&in).Encode(x); err != nil {
		t.Fatal("write input:", err)
	}
	// Evaluate function.
	var out bytes.Buffer
	if err := ReadEvalWrite(&out, sqr, &in); err != nil {
		t.Fatal(err)
	}
	// Read output from stream.
	var y float64
	if err := json.NewDecoder(&out).Decode(&y); err != nil {
		t.Fatal(err)
	}
	checkEqual(t, 25, y)
}

// Evaluate a function which encounters an error.
// Pass arugments and receive result by JSON.
func TestReadEvalWrite_error(t *testing.T) {
	x := float64(-1)
	// Write input to stream.
	var in bytes.Buffer
	if err := json.NewEncoder(&in).Encode(x); err != nil {
		t.Fatal("write input:", err)
	}
	// Evaluate function.
	var out bytes.Buffer
	if err := ReadEvalWrite(&out, sqrt, &in); err == nil {
		t.Fatal("expected to receive error")
	}
}

// Compare two floating-point numbers.
func checkEqual(t *testing.T, want, got float64) {
	const eps = 1e-9
	if math.Abs(want-got) > eps {
		t.Fatalf("want %.6g, got %.6g", want, got)
	}
}
