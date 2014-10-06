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

// If n is greater than 1, the -J argument is supplied.
func submit(n int, jobargs []string, name, dir, userargs string, subout, suberr io.Writer) error {
	var args []string
	// Set task name.
	args = append(args, "-N", name)
	// Set number of jobs.
	if n > 1 {
		// TODO: Handle map of 1 task.
		args = append(args, "-J", fmt.Sprintf("1-%d", n))
	}
	// Put stdout and stderr in temporary dir.
	args = append(args, "-e", path.Clean(dir)+"/")
	args = append(args, "-o", path.Clean(dir)+"/")
	// Wait for all jobs to finish.
	args = append(args, "-W", "block=TRUE,sandbox=PRIVATE")
	// Use same environment variables.
	args = append(args, "-V")
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
	if len(jobargs) > 0 {
		args = append(args, jobargs...)
	}

	cmd := exec.Command("qsub", args...)
	// Re-route stdout and stderr.
	cmd.Stdout = subout
	cmd.Stderr = suberr
	log.Printf("qsub arguments: %#v", args)
	return cmd.Run()
}
