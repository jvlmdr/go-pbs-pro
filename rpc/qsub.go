package grideng

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func submit(n int, resources string, cmdArgs []string) error {
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
	if len(resources) > 0 {
		args = append(args, "-l", resources)
	}
	// Redirect stdout.
	args = append(args, "-o", `stdout-$TASK_ID`)
	// Redirect stderr.
	args = append(args, "-e", `stderr-$TASK_ID`)

	// Name of executable to run.
	args = append(args, os.Args[0])
	args = append(args, cmdArgs...)

	// Submit.
	cmd := exec.Command("qsub", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var b bytes.Buffer
	fmt.Fprint(&b, "qsub")
	for _, arg := range args {
		fmt.Fprint(&b, " ", arg)
	}
	log.Print(b.String())

	return cmd.Run()
}
