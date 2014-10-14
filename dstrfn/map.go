package dstrfn

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"

	"github.com/jvlmdr/go-file/fileutil"
)

func MapFunc(f string, y, x interface{}, p ...interface{}) error {
	return Map(f, y, x, p, DefaultStdout, DefaultStderr, nil)
}

// Map computes y[i] = f(x[i], p) for all i.
//
// The function is specified by name and must already be registered.
// The input x must be a slice or similar (see reflect.Value.Index()).
// The output y must be a pointer to a slice.
// If the length of y is sufficient to hold the output, it will be over-written.
// If it is not sufficient, a new array will be allocated.
// After a succesful call, the length of y will match that of x.
func Map(f string, y, x, p interface{}, stdout, stderr io.Writer, flags []string) error {
	task, there := mapTasks[f]
	if !there {
		return fmt.Errorf(`map task not found: "%s"`, f)
	}

	// Recursively invoked closure.
	var do func(task *mapTaskSpec, y, x interface{}, chunk bool) (string, error)
	do = func(task *mapTaskSpec, y, x interface{}, chunk bool) (string, error) {
		n := reflect.ValueOf(x).Len()
		y = ensureLenAndDeref(y, n)
		// y now has correct len, is not a pointer, and can be modified.

		if chunk {
			u, inds := split(x, 1, max(task.ChunkLen, 1))
			// Create slice of slices for output.
			vtyp := reflect.SliceOf(reflect.TypeOf(y))
			v := reflect.New(vtyp).Interface()
			dir, err := do(task, v, u, false)
			v = deref(v)
			if err != nil {
				mapErr := err.(MapError)
				// Need to re-map task errors.
				taskErrs := make(map[int]error)
				for i := range inds {
					if err := mapErr.Tasks[i]; err != nil {
						// Give error to all members.
						for _, p := range inds[i] {
							taskErrs[p] = err
						}
						continue
					}
					// No error occured. Move outputs.
					for j, p := range inds[i] {
						vij := reflect.ValueOf(v).Index(i).Index(j)
						yp := reflect.ValueOf(y).Index(p)
						yp.Set(vij)
					}
				}
				return dir, MapError{mapErr.Master, taskErrs, n}
			}
			mergeTo(y, v)
			return dir, nil
		}

		// Create temporary directory.
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dir, err := ioutil.TempDir(wd, f+"-")
		if err != nil {
			return "", err
		}

		// Save each input to file.
		xval := reflect.ValueOf(x)
		for i := 0; i < xval.Len(); i++ {
			inFile := path.Join(dir, fmt.Sprintf("in-%d.json", i))
			err := fileutil.SaveExt(inFile, xval.Index(i).Interface())
			if err != nil {
				return dir, fmt.Errorf("save input %d: %v", i, err)
			}
		}
		if p != nil {
			confFile := path.Join(dir, "conf.json")
			err := fileutil.SaveExt(confFile, p)
			if err != nil {
				return dir, fmt.Errorf("save config: %v", err)
			}
		}

		// Invoke qsub.
		jobargs := []string{"-dstrfn.task", f, "-dstrfn.map", fmt.Sprint(n), "-dstrfn.dir", dir}
		if len(flags) > 0 {
			jobargs = append(jobargs, flags...)
		}
		execErr, err := submit(n, jobargs, f, dir, task.Flags, nil, nil)
		if err != nil {
			return dir, err
		}

		taskErrs := make(map[int]error)
		for i := 0; i < n; i++ {
			// Load from output file.
			outFile := path.Join(dir, fmt.Sprintf("out-%d.json", i))
			errFile := path.Join(dir, fmt.Sprintf("err-%d.json", i))
			yi := reflect.ValueOf(y).Index(i).Addr().Interface()

			if _, err := os.Stat(outFile); err == nil {
				// If output file exists, attempt to load.
				if err := fileutil.LoadExt(outFile, yi); err != nil {
					taskErrs[i] = fmt.Errorf("load output: %v", err)
					continue
				}
			} else if !os.IsNotExist(err) {
				// Could not stat file.
				taskErrs[i] = err
				continue
			} else {
				// Output file did not exist. Try to load error file.
				if _, err := os.Stat(errFile); err == nil {
					// Error file exists. Attempt to load.
					var str string
					if err := fileutil.LoadExt(errFile, &str); err != nil {
						taskErrs[i] = err
						continue
					}
					taskErrs[i] = errors.New(str)
					continue
				} else if !os.IsNotExist(err) {
					// Could not stat file.
					taskErrs[i] = err
					continue
				}
				taskErrs[i] = fmt.Errorf("could not find output or error files: job %d", i)
				continue
			}
		}

		if execErr != nil {
			return dir, MapError{execErr, taskErrs, n}
		}
		if len(taskErrs) > 0 {
			return dir, MapError{Tasks: taskErrs, Len: n}
		}
		return dir, nil
	}

	tmpdir, err := do(task, y, x, task.Chunk)
	if err != nil {
		return err
	}
	// Only remove temporary directory if there was no error.
	if !debug {
		if err := removeAll(tmpdir); err != nil {
			log.Println(err)
		}
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
