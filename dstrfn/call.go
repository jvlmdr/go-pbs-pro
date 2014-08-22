package dstrfn

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/jvlmdr/go-file/fileutil"
)

var (
	DefaultStdout = os.Stderr
	DefaultStderr = os.Stderr
)

func CallFunc(f string, y interface{}, x ...interface{}) error {
	if len(x) == 1 {
		return Call(f, y, x[0], DefaultStdout, DefaultStderr)
	}
	return Call(f, y, x, DefaultStdout, DefaultStderr)
}

// Call calls the function and saves the output to the specified file.
// It does not load the result into memory.
// If the file already exists, it does not call the function.
func Call(f string, y, x interface{}, stdout, stderr io.Writer) error {
	task, there := tasks[f]
	if !there {
		return fmt.Errorf(`task not found: "%s"`, f)
	}

	// Create temporary directory.
	dir, err := ioutil.TempDir(".", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	inFile := path.Join(dir, "in.json")
	outFile := path.Join(dir, "out.json")
	errFile := path.Join(dir, "err.json")
	// Save input.
	if err := fileutil.SaveJSON(inFile, x); err != nil {
		return err
	}

	// Invoke qsub.
	jobargs := []string{"-dstrfn.task", f, "-dstrfn.dir", dir}
	err = submit(1, jobargs, f, task.Flags, stdout, stderr, task.Stdout, task.Stderr)
	if err != nil {
		return err
	}

	if _, err := os.Stat(errFile); err == nil {
		// Error file exists. Attempt to load.
		var str string
		if err := fileutil.LoadExt(errFile, &str); err != nil {
			return fmt.Errorf("load error file: %v", err)
		}
		return errors.New(str)
	} else if !os.IsNotExist(err) {
		// Could not stat file.
		return fmt.Errorf("stat error file: %v", err)
	}
	// Error file does not exist.

	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		return errors.New("could not find output or error files")
	} else if err != nil {
		return fmt.Errorf("stat output file: %v", err)
	}
	// Output file exists.

	if y == nil {
		return nil
	}
	if err := fileutil.LoadExt(outFile, y); err != nil {
		return err
	}
	return nil
}
