// +build debug

package main

import (
	"fmt"
	"os"
	"runtime/debug"
)

func init() {
	// Capture panics and print stack trace before crashing
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "PANIC: %v\n", r)
			fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", debug.Stack())
			os.Exit(1)
		}
	}()
}
