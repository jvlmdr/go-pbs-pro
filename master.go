package grideng

import (
	"fmt"
	"os"
	"os/exec"
)

// Executes all tasks and returns a list of output files.
func Master(inputs InputList, resources string, cmdArgs []string) ([]string, error) {
	// Serialize inputs.
	if err := saveAllInputs(inputs); err != nil {
		return nil, err
	}

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
	args = append(args, "-t", fmt.Sprintf("1-%d", inputs.Len()))
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

	fmt.Print("qsub")
	for _, arg := range args {
		fmt.Print(" ", arg)
	}
	fmt.Println()

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Check success/failure of tasks.
	files := make([]string, inputs.Len())
	for i := 0; i < inputs.Len(); i++ {
		name := inputs.At(i).Name()
		outfile := outputFile(name)
		// Check if output file exists.
		if _, err := os.Stat(outfile); err == nil {
			files[i] = outfile
		}
	}
	return files, nil
}

// Attempts to save each task input to a file.
func saveAllInputs(inputs InputList) error {
	for i := 0; i < inputs.Len(); i++ {
		// Grid Engine task ID (one-indexed not zero-indexed).
		num := i + 1
		// Save input to file.
		file := inputFile(num)
		if err := saveInput(inputs.At(i), file); err != nil {
			return err
		}
	}
	return nil
}

// Attempts to save the input for a task to a file.
// Open files are closed on return.
func saveInput(input Input, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := input.Write(file); err != nil {
		return err
	}
	return nil
}

// Reads all files into the list of outputs.
func LoadOutputs(outputs OutputList, files []string) error {
	for i, file := range files {
		if err := readOutput(outputs, i, file); err != nil {
			return err
		}
	}
	return nil
}

// Attempts to load the output for a task from a file.
// Open files are closed on return.
func readOutput(outputs OutputList, i int, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := outputs.Read(i, file); err != nil {
		return err
	}
	return nil
}
