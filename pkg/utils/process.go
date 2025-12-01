package utils

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// ProcessResult holds the result of a process execution
type ProcessResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// RunProcess executes a process with timeout
func RunProcess(ctx context.Context, name string, args ...string) (*ProcessResult, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, name, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	stdoutBytes, _ := io.ReadAll(stdout)
	stderrBytes, _ := io.ReadAll(stderr)

	err = cmd.Wait()

	result := &ProcessResult{
		Stdout:   string(stdoutBytes),
		Stderr:   string(stderrBytes),
		Duration: time.Since(start),
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else if err != nil {
		return result, err
	}

	return result, nil
}

// RunProcessWithTimeout executes a process with a specific timeout
func RunProcessWithTimeout(timeout time.Duration, name string, args ...string) (*ProcessResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return RunProcess(ctx, name, args...)
}

// RunProcessWithStdin executes a process with stdin input
func RunProcessWithStdin(ctx context.Context, input []byte, name string, args ...string) (*ProcessResult, error) {
	start := time.Now()

	cmd := exec.CommandContext(ctx, name, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	// Write input to stdin
	go func() {
		defer stdin.Close()
		stdin.Write(input)
	}()

	stdoutBytes, _ := io.ReadAll(stdout)
	stderrBytes, _ := io.ReadAll(stderr)

	err = cmd.Wait()

	result := &ProcessResult{
		Stdout:   string(stdoutBytes),
		Stderr:   string(stderrBytes),
		Duration: time.Since(start),
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else if err != nil {
		return result, err
	}

	return result, nil
}

// IsProcessRunning checks if a process with given name is running
func IsProcessRunning(name string) bool {
	cmd := exec.Command("pgrep", "-x", name)
	err := cmd.Run()
	return err == nil
}

// CheckServiceAvailable checks if a service is responding on given URL
func CheckServiceAvailable(url string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Simple HTTP check using curl
	cmd := exec.CommandContext(ctx, "curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", url)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if we got a valid HTTP response
	return len(output) > 0 && output[0] >= '2' && output[0] <= '3'
}
