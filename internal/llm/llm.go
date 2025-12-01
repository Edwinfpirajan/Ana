// Package llm provides language model integration for AnaStreamer
package llm

import (
	"context"
	"fmt"

	"github.com/anastreamer/ana/internal/config"
)

// Action represents a parsed action from the LLM
type Action struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Reply  string                 `json:"reply"`
}

// IsEmpty returns true if the action is empty/invalid
func (a Action) IsEmpty() bool {
	return a.Action == ""
}

// GetStringParam gets a string parameter from the action
func (a Action) GetStringParam(key string) string {
	if v, ok := a.Params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetIntParam gets an int parameter from the action
func (a Action) GetIntParam(key string) int {
	if v, ok := a.Params[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case int64:
			return int(val)
		}
	}
	return 0
}

// GetFloatParam gets a float64 parameter from the action
func (a Action) GetFloatParam(key string) float64 {
	if v, ok := a.Params[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

// GetBoolParam gets a bool parameter from the action
func (a Action) GetBoolParam(key string) bool {
	if v, ok := a.Params[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// Provider is the interface for LLM providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Complete sends a prompt to the LLM and returns an Action
	Complete(ctx context.Context, prompt string) (Action, error)

	// CompleteRaw sends a prompt and returns the raw response
	CompleteRaw(ctx context.Context, prompt string) (string, error)

	// IsAvailable checks if the provider is available
	IsAvailable(ctx context.Context) bool

	// Close releases any resources
	Close() error
}

// New creates a new LLM provider based on configuration
func New(cfg *config.Config) (Provider, error) {
	switch cfg.LLM.Provider {
	case "ollama":
		return NewOllamaProvider(cfg.LLM.Ollama)
	case "openai":
		return NewOpenAIProvider(cfg.LLM.OpenAI)
	case "auto":
		return NewAutoProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", cfg.LLM.Provider)
	}
}
