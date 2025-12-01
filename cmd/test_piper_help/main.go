package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	workDir, _ := os.Getwd()
	binaryPath := filepath.Join(workDir, "bin", "piper", "piper.exe")

	fmt.Printf("Testing: %s --help\n", binaryPath)

	cmd := exec.Command(binaryPath, "--help")
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	fmt.Printf("Exit code: %d (0x%x)\n", exitCode, uint32(exitCode))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	if stdout.Len() > 0 {
		fmt.Printf("STDOUT:\n%s\n", stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Printf("STDERR:\n%s\n", stderr.String())
	}
}
