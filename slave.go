package grideng

import (
	"log"
	"os"
)

// Loads, executes and saves a task.
func Slave(num int, reader InputReader) {
	if err := slave(num, reader); err != nil {
		panic(err)
	}
}

// Loads, executes and saves a task.
func slave(num int, reader InputReader) error {
	// Load task input from file.
	infile := inputFile(num)
	input, err := loadInput(reader, infile)
	if err != nil {
		return err
	}

	name := input.Name()
	outfile := outputFile(name)

	// Check if output file already exists.
	if _, err := os.Stat(outfile); err == nil {
		log.Println("Skipping task")
		return nil
	}

	// Do the thing.
	output, err := input.Execute()
	if err != nil {
		return err
	}

	// Save output.
	if err := saveOutput(output, outfile); err != nil {
		return err
	}
	return nil
}

// Attempts to load an unexecuted task from an input file.
// Open files are closed on return.
func loadInput(reader InputReader, name string) (Input, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	input, err := reader.Read(file)
	if err != nil {
		return nil, err
	}
	return input, nil
}

func saveOutput(output Output, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := output.Write(file); err != nil {
		return err
	}
	return nil
}
