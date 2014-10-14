package dstrfn

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/jvlmdr/go-file/fileutil"
)

var (
	DefaultStdout = os.Stderr
	DefaultStderr = os.Stderr
)

func CallFunc(f string, y interface{}, x ...interface{}) error {
	if len(x) == 1 {
		return Call(f, y, x[0], DefaultStdout, DefaultStderr, nil)
	}
	return Call(f, y, x, DefaultStdout, DefaultStderr, nil)
}

func Args(x ...interface{}) []interface{} {
	return x
}

// Call calls the function and saves the output to the specified file.
// It does not load the result into memory.
// If the file already exists, it does not call the function.
func Call(f string, y, x interface{}, stdout, stderr io.Writer, flags []string) error {
	task, there := tasks[f]
	if !there {
		return fmt.Errorf(`task not found: "%s"`, f)
	}

	// Create temporary directory.
	dir, err := ioutil.TempDir(".", f+"-")
	if err != nil {
		return err
	}

	inFile := path.Join(dir, "in.json")
	outFile := path.Join(dir, "out.json")
	errFile := path.Join(dir, "err.json")
	// Save input.
	if err := fileutil.SaveJSON(inFile, x); err != nil {
		return err
	}

	// Invoke qsub.
	jobargs := []string{"-dstrfn.task", f, "-dstrfn.dir", dir}
	if len(flags) > 0 {
		jobargs = append(jobargs, flags...)
	}
	execErr, err := submit(1, jobargs, f, dir, task.Flags, stdout, stderr)
	if err != nil {
		return err
	}
	if execErr != nil {
		return execErr
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

	if y != nil {
		// Output required.
		if _, err := os.Stat(outFile); os.IsNotExist(err) {
			return errors.New("could not find output or error files")
		} else if err != nil {
			return fmt.Errorf("stat output file: %v", err)
		}
		if err := fileutil.LoadExt(outFile, y); err != nil {
			return err
		}
	}
	// Only remove temporary directory if there was no error.
	if !debug {
		return removeAll(dir)
	}
	return nil
}

func removeAll(fname string) error {
	for {
		err := os.RemoveAll(fname)
		if err == nil {
			return nil
		}
		log.Print(err)
		time.Sleep(1)
	}
}
