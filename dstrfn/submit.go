package dstrfn

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
)

func submit(n int, userargs, jobargs []string, name string, subout, suberr io.Writer, jobout, joberr bool) error {
	var args []string
	// Set task name.
	args = append(args, "-N", name)
	// Set number of jobs.
	if n > 1 {
		args = append(args, "-J", fmt.Sprintf("1-%d", n))
	}
	// Wait for all jobs to finish.
	args = append(args, "-Wblock=true")
	// Use same environment variables.
	args = append(args, "-V")
	// Where to send stdout and stderr.
	switch {
	case jobout && joberr:
		args = append(args, "-k", "n")
	case jobout:
		args = append(args, "-k", "e")
	case joberr:
		args = append(args, "-k", "o")
	default:
		args = append(args, "-k", "oe")
	}
	// Set resources.
	if len(userargs) > 0 {
		args = append(args, userargs...)
	}

	// Name of executable to run.
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	args = append(args, "--", path.Join(wd, os.Args[0]))
	args = append(args, jobargs...)

	// Submit.
	cmd := exec.Command("qsub", args...)
	// Do not pipe stdout to stdout.
	cmd.Stdout = subout
	cmd.Stderr = suberr

	var b bytes.Buffer
	fmt.Fprint(&b, "qsub")
	for _, arg := range args {
		fmt.Fprint(&b, " ", arg)
	}
	log.Println("invoke:", b.String())

	return cmd.Run()
}
