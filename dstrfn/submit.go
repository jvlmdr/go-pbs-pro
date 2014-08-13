package dstrfn

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func submit(isMap bool, n int, jobargs []string, name, userargs string, subout, suberr io.Writer, jobout, joberr bool) error {
	var args []string
	// Set task name.
	args = append(args, "-N", name)
	// Set number of jobs.
	if isMap {
		// TODO: Handle map of 1 task.
		args = append(args, "-J", fmt.Sprintf("1-%d", n))
	}
	// Wait for all jobs to finish.
	args = append(args, "-Wblock=true")
	// Use same environment variables.
	args = append(args, "-V")
	// Where to send stdout and stderr.
	args = append(args, "-k", keepStr(jobout, joberr))
	// Set resources.
	if len(userargs) > 0 {
		args = append(args, strings.Split(userargs, " ")...)
	}

	// Full path of executable to run.
	self := os.Args[0]
	if !path.IsAbs(self) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		self = path.Join(wd, os.Args[0])
	}
	args = append(args, "--", self)
	args = append(args, jobargs...)

	cmd := exec.Command("qsub", args...)
	// Re-route stdout and stderr.
	cmd.Stdout = subout
	cmd.Stderr = suberr
	log.Printf("qsub arguments: %#v", args)
	return cmd.Run()
}

func keepStr(out, err bool) string {
	switch {
	case out && err:
		return "n"
	case out:
		return "e"
	case err:
		return "o"
	default:
		return "oe"
	}
}
