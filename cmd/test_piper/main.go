package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Get absolute paths
	workDir, _ := os.Getwd()
	fmt.Printf("Working directory: %s\n", workDir)

	binaryPath := filepath.Join(workDir, "bin", "piper", "piper.exe")
	modelPath := filepath.Join(workDir, "assets", "voices", "piper", "es_ES-davefx-medium.onnx")
	outputPath := filepath.Join(os.TempDir(), "test_piper.wav")

	fmt.Printf("Binary: %s\n", binaryPath)
	fmt.Printf("Model: %s\n", modelPath)
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Println("---")

	// Check if files exist
	if _, err := os.Stat(binaryPath); err != nil {
		fmt.Printf("Binary check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Binary exists")

	if _, err := os.Stat(modelPath); err != nil {
		fmt.Printf("Model check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Model exists")

	// Create command
	args := []string{
		"--model", modelPath,
		"--output_file", outputPath,
	}

	fmt.Printf("Running: %s %v\n", binaryPath, args)

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

	// Start command
	fmt.Println("Starting command...")
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start command: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Command started")

	// Write text
	fmt.Println("Writing text to stdin...")
	if _, err := stdin.Write([]byte("Hola mundo\n")); err != nil {
		fmt.Printf("Failed to write to stdin: %v\n", err)
		os.Exit(1)
	}
	stdin.Close()
	fmt.Println("✓ Text written and stdin closed")

	// Wait for completion
	fmt.Println("Waiting for command to complete...")
	err = cmd.Wait()
	exitCode := cmd.ProcessState.ExitCode()
	fmt.Printf("Exit code: %d\n", exitCode)

	if err != nil {
		fmt.Printf("Command error: %v\n", err)
	}

	// Print output
	if stdout.Len() > 0 {
		fmt.Printf("STDOUT:\n%s\n", stdout.String())
	}
	if stderr.Len() > 0 {
		fmt.Printf("STDERR:\n%s\n", stderr.String())
	}

	// Check if output file was created
	if _, err := os.Stat(outputPath); err == nil {
		fileInfo, _ := os.Stat(outputPath)
		fmt.Printf("\n✓ Audio file created: %d bytes\n", fileInfo.Size())
	} else {
		fmt.Printf("\n✗ Audio file NOT created\n")
	}
}
