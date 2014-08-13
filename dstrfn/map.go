package dstrfn

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"

	"github.com/jvlmdr/go-file/fileutil"
)

func MapFunc(f string, y, x interface{}, p ...interface{}) error {
	return Map(f, y, x, p, DefaultStdout, DefaultStderr)
}

// Map computes y[i] = f(x[i], p) for all i.
//
// The function is specified by name and must already be registered.
// The input x must be a slice or similar (see reflect.Value.Index()).
// The output y must be a pointer to a slice.
// If the length of y is sufficient to hold the output, it will be over-written.
// If it is not sufficient, a new array will be allocated.
// After a succesful call, the length of y will match that of x.
func Map(f string, y, x, p interface{}, stdout, stderr io.Writer) error {
	task, there := mapTasks[f]
	if !there {
		return fmt.Errorf(`map task not found: "%s"`, f)
	}

	var do func(task *mapTaskSpec, y, x interface{}, chunk bool) error
	do = func(task *mapTaskSpec, y, x interface{}, chunk bool) error {
		n := reflect.ValueOf(x).Len()
		y = ensureLenAndDeref(y, n)

		if chunk {
			u := split(x, 1, max(task.ChunkLen, 1))
			// Create slice of slices for output.
			vtyp := reflect.SliceOf(reflect.TypeOf(y))
			v := reflect.New(vtyp).Interface()
			err := do(task, v, u, false)
			if err != nil {
				return err
			}
			mergeTo(y, deref(v))
			return nil
		}

		// Create temporary directory.
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		tmpDir, err := ioutil.TempDir(wd, "")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		// Save each input to file.
		xval := reflect.ValueOf(x)
		for i := 0; i < xval.Len(); i++ {
			inFile := path.Join(tmpDir, fmt.Sprintf("in-%d.json", i))
			err := fileutil.SaveExt(inFile, xval.Index(i).Interface())
			if err != nil {
				return err
			}
		}
		if p != nil {
			confFile := path.Join(tmpDir, "conf.json")
			err := fileutil.SaveExt(confFile, p)
			if err != nil {
				return err
			}
		}

		// Invoke qsub.
		jobargs := []string{"-dstrfn.task", f, "-dstrfn.map", "-dstrfn.dir", tmpDir}
		err = submit(true, n, jobargs, f, task.Flags, nil, nil, task.Stdout, task.Stderr)
		if err != nil {
			return err
		}

		for i := 0; i < n; i++ {
			// Load from output file.
			outFile := path.Join(tmpDir, fmt.Sprintf("out-%d.json", i))
			errFile := path.Join(tmpDir, fmt.Sprintf("err-%d.json", i))
			yi := reflect.ValueOf(y).Index(i).Addr().Interface()

			if _, err := os.Stat(outFile); err == nil {
				// If output file exists, attempt to load.
				if err := fileutil.LoadExt(outFile, yi); err != nil {
					return err
				}
			} else if !os.IsNotExist(err) {
				// Could not stat file.
				return err
			} else {
				// Output file did not exist. Try to load error file.
				if _, err := os.Stat(errFile); err == nil {
					// Error file exists. Attempt to load.
					var str string
					if err := fileutil.LoadExt(errFile, &str); err != nil {
						return err
					}
					return errors.New(str)
				} else if !os.IsNotExist(err) {
					// Could not stat file.
					return err
				}
				return fmt.Errorf("could not find output or error files: job %d", i)
			}
		}
		return nil
	}

	if err := do(task, y, x, task.Chunk); err != nil {
		return err
	}
	return nil
}

// Ensures that dst has length n and then de-references the pointer.
// The slice header is sufficient to change the underlying elements.
func ensureLenAndDeref(dst interface{}, n int) interface{} {
	// De-reference pointer.
	val := reflect.ValueOf(dst).Elem()
	if val.Len() >= n {
		return val.Slice(0, n).Interface()
	}
	// Not big enough, re-allocate.
	val.Set(reflect.MakeSlice(val.Type(), n, n))
	return val.Interface()
}
