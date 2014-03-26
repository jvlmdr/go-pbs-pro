package grideng

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func submit(n int, res string, cmdArgs []string, stdout, stderr io.Writer) error {
	var args []string
	// Submitting a binary job.
	args = append(args, "-b", "y")
	// Wait for jobs to finish.
	args = append(args, "-sync", "y")
	// Use current working directory.
	args = append(args, "-cwd")
	// Use same environment variables.
	args = append(args, "-V")
	// Use same environment variables.
	args = append(args, "-t", fmt.Sprintf("1-%d", n))
	// Set resources.
	if len(res) > 0 {
		args = append(args, "-l", res)
	}

	// Name of executable to run.
	args = append(args, os.Args[0])
	args = append(args, cmdArgs...)

	// Submit.
	cmd := exec.Command("qsub", args...)
	// Do not pipe stdout to stdout.
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	var b bytes.Buffer
	fmt.Fprint(&b, "qsub")
	for _, arg := range args {
		fmt.Fprint(&b, " ", arg)
	}
	log.Print(b.String())

	return cmd.Run()
}
