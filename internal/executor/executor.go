// Package executor provides action executors for AnaStreamer
package executor

import (
	"context"
	"fmt"

	"github.com/anastreamer/ana/internal/llm"
)

// Result holds the result of an action execution
type Result struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// NewResult creates a successful result
func NewResult(message string) Result {
	return Result{
		Success: true,
		Message: message,
	}
}

// NewResultWithData creates a successful result with data
func NewResultWithData(message string, data map[string]interface{}) Result {
	return Result{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResult creates an error result
func NewErrorResult(err error) Result {
	return Result{
		Success: false,
		Error:   err.Error(),
	}
}

// Executor is the interface for action executors
type Executor interface {
	// Name returns the executor name (e.g., "twitch", "obs", "music")
	Name() string

	// SupportedActions returns the list of actions this executor handles
	SupportedActions() []string

	// CanHandle returns true if this executor can handle the given action
	CanHandle(action string) bool

	// Execute executes an action
	Execute(ctx context.Context, action llm.Action) (Result, error)

	// IsAvailable checks if the executor is ready to handle actions
	IsAvailable() bool

	// Close releases any resources
	Close() error
}

// StatusProvider is an optional interface for executors that provide detailed status
type StatusProvider interface {
	// GetStatus returns detailed status information
	GetStatus() string
}

// Registry holds all registered executors
type Registry struct {
	executors map[string]Executor
}

// NewRegistry creates a new executor registry
func NewRegistry() *Registry {
	return &Registry{
		executors: make(map[string]Executor),
	}
}

// Register adds an executor to the registry
func (r *Registry) Register(exec Executor) {
	r.executors[exec.Name()] = exec
}

// Get returns an executor by name
func (r *Registry) Get(name string) (Executor, bool) {
	exec, ok := r.executors[name]
	return exec, ok
}

// FindExecutor finds the executor that can handle the given action
func (r *Registry) FindExecutor(action string) (Executor, error) {
	for _, exec := range r.executors {
		if exec.CanHandle(action) {
			return exec, nil
		}
	}
	return nil, fmt.Errorf("no executor found for action: %s", action)
}

// Execute finds the appropriate executor and executes the action
func (r *Registry) Execute(ctx context.Context, action llm.Action) (Result, error) {
	exec, err := r.FindExecutor(action.Action)
	if err != nil {
		return NewErrorResult(err), err
	}

	if !exec.IsAvailable() {
		err := fmt.Errorf("executor %s is not available", exec.Name())
		return NewErrorResult(err), err
	}

	return exec.Execute(ctx, action)
}

// GetAllActions returns all supported actions from all executors
func (r *Registry) GetAllActions() []string {
	var actions []string
	for _, exec := range r.executors {
		actions = append(actions, exec.SupportedActions()...)
	}
	return actions
}

// Close closes all executors
func (r *Registry) Close() error {
	var lastErr error
	for _, exec := range r.executors {
		if err := exec.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
