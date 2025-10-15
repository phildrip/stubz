package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/phildrip/toe/options"
)

func run(stdout, stderr io.Writer, args []string) int {
	var stubDirFlag string
	var outputFile string // Keep outputFile as a flag

	fs := flag.NewFlagSet("toe", flag.ContinueOnError)
	fs.SetOutput(stderr) // Direct flag errors to stderr
	fs.StringVar(&stubDirFlag,
		"stub-dir",
		"",
		"generate stub in a specific subdirectory (e.g., \"stubs\") and use its name as package")
	fs.StringVar(&outputFile,
		"o",
		"",
		"output file name (if not provided, defaults to stub_<interface-lowercased>.go in the default/specified stub-dir)")

	// Parse command-line arguments, excluding the program name
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}

	if fs.NArg() != 2 {
		fmt.Fprintf(stderr,
			"Usage: %s [-stub-dir <dir>] [-o <output.go>] <input_directory> <interface>\n",
			args[0])
		return 1
	}

	// Get inputDir and interfaceName now that flags are parsed
	inputDir := fs.Arg(0)
	interfaceName := fs.Arg(1)

	// Determine the stub directory and filename
	actualStubDir := stubDirFlag
	if actualStubDir == "" {
		actualStubDir = "stubs"
	}

	outputFilename := fmt.Sprintf("stub_%s.go", strings.ToLower(interfaceName))

	// If -o is not provided, derive it from stubDir and interface name
	if outputFile == "" {
		outputFile = filepath.Join(actualStubDir, outputFilename)
	}

	// Ensure the output directory exists
	// This is handled by main_test.go for tests. For direct CLI usage, it's needed.
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		fmt.Fprintf(stderr, "Error creating output directory for %s: %v\n", outputFile, err)
		return 1
	}

	var opts = &options.StubOptions{WithLocking: true}

	interfaceData, err := FindInterface(inputDir,
		interfaceName,
		actualStubDir)
	if err != nil {
		fmt.Fprintf(stderr, "Error finding interface: %v\n", err)
		return 1
	}

	stubCode, err := GenerateStubCode(interfaceData, opts)
	if err != nil {
		fmt.Fprintf(stderr, "Error generating stub: %v\n", err)
		return 1
	}

	// Write the generated code (already formatted by the generator)
	err = os.WriteFile(outputFile, []byte(stubCode), 0644)
	if err != nil {
		fmt.Fprintf(stderr, "Error writing output file: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Stub generated in %s\n", outputFile)

	return 0
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args))
}
