package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/anastreamer/ana/internal/config"
)

// autoProvider switches between available LLM vendors (local/openai) at runtime.
type autoProvider struct {
	providers []Provider
}

// NewAutoProvider instantiates the auto LLM provider.
func NewAutoProvider(cfg *config.Config) (Provider, error) {
	var providers []Provider
	var errs []string

	streamerName := cfg.General.StreamerName
	if p, err := NewOllamaProvider(cfg.LLM.Ollama, streamerName); err == nil {
		providers = append(providers, p)
	} else {
		errs = append(errs, fmt.Sprintf("ollama: %v", err))
	}

	if p, err := NewOpenAIProvider(cfg.LLM.OpenAI, streamerName); err == nil {
		providers = append(providers, p)
	} else {
		errs = append(errs, fmt.Sprintf("openai: %v", err))
	}

	if len(providers) == 0 {
		msg := "auto LLM provider requires at least one valid backend"
		if len(errs) > 0 {
			msg = fmt.Sprintf("%s (%s)", msg, strings.Join(errs, "; "))
		}
		return nil, fmt.Errorf(msg)
	}

	return &autoProvider{providers: providers}, nil
}

func (a *autoProvider) Name() string {
	names := make([]string, len(a.providers))
	for i, p := range a.providers {
		names[i] = p.Name()
	}
	return fmt.Sprintf("auto(%s)", strings.Join(names, "+"))
}

func (a *autoProvider) Complete(ctx context.Context, prompt string) (Action, error) {
	p, err := a.selectProvider(ctx)
	if err != nil {
		return Action{}, err
	}
	return p.Complete(ctx, prompt)
}

func (a *autoProvider) CompleteRaw(ctx context.Context, prompt string) (string, error) {
	p, err := a.selectProvider(ctx)
	if err != nil {
		return "", err
	}
	return p.CompleteRaw(ctx, prompt)
}

func (a *autoProvider) IsAvailable(ctx context.Context) bool {
	for _, p := range a.providers {
		if p.IsAvailable(ctx) {
			return true
		}
	}
	return false
}

func (a *autoProvider) Close() error {
	var errs []string
	for _, p := range a.providers {
		if err := p.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing LLM providers: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (a *autoProvider) selectProvider(ctx context.Context) (Provider, error) {
	for _, p := range a.providers {
		if p.IsAvailable(ctx) {
			return p, nil
		}
	}

	if len(a.providers) == 0 {
		return nil, fmt.Errorf("no LLM providers configured")
	}
	return a.providers[0], nil
}
