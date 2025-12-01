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
	modelPath := filepath.Join(workDir, "assets", "voices", "piper", "es_ES-davefx-medium.onnx")
	outputPath := filepath.Join(os.TempDir(), "test_piper_debug.wav")

	fmt.Printf("Binary: %s\n", binaryPath)
	fmt.Printf("Model: %s\n", modelPath)
	fmt.Printf("Output: %s\n", outputPath)

	// Try with --debug and --espeak_data
	espeak_data := filepath.Join(workDir, "bin", "piper", "espeak-ng-data")

	args := []string{
		"--model", modelPath,
		"--output_file", outputPath,
		"--espeak_data", espeak_data,
		"--debug",
	}

	fmt.Printf("\nRunning with args: %v\n\n", args)

	cmd := exec.Command(binaryPath, args...)
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	// Get stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Failed to get stdin pipe: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start command: %v\n", err)
		os.Exit(1)
	}

	// Write text
	fmt.Println("Writing text...")
	stdin.Write([]byte("Hola mundo\n"))
	stdin.Close()

	// Wait
	err = cmd.Wait()
	exitCode := cmd.ProcessState.ExitCode()

	fmt.Printf("\nExit code: %d (0x%x)\n", exitCode, uint32(exitCode))

	if stdout.Len() > 0 {
		fmt.Printf("STDOUT:\n%s\n", stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Printf("STDERR:\n%s\n", stderr.String())
	}

	if _, err := os.Stat(outputPath); err == nil {
		fileInfo, _ := os.Stat(outputPath)
		fmt.Printf("✓ Success! Audio file: %d bytes\n", fileInfo.Size())
	} else {
		fmt.Printf("✗ Failed: No audio file created\n")
	}
}
