package tts

import (
	"context"
	"fmt"
	"strings"

	"github.com/anastreamer/ana/internal/config"
)

// autoProvider tries available TTS vendors (local/cloud) in order.
type autoProvider struct {
	providers []Provider
}

// NewAutoProvider builds the auto TTS provider.
func NewAutoProvider(cfg *config.Config) (Provider, error) {
	var providers []Provider
	var errs []string

	// Try Windows TTS first (always available on Windows)
	if p, err := NewWindowsTTSProvider(cfg.TTS.OpenAI); err == nil {
		providers = append(providers, p)
	} else {
		errs = append(errs, fmt.Sprintf("windows-tts: %v", err))
	}

	if p, err := NewPiperProvider(cfg.TTS.Piper); err == nil {
		providers = append(providers, p)
	} else {
		errs = append(errs, fmt.Sprintf("piper: %v", err))
	}

	if p, err := NewOpenAIProvider(cfg.TTS.OpenAI); err == nil {
		providers = append(providers, p)
	} else {
		errs = append(errs, fmt.Sprintf("openai: %v", err))
	}

	if len(providers) == 0 {
		msg := "auto TTS provider requires at least one backend"
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

func (a *autoProvider) Speak(ctx context.Context, text string) error {
	p, err := a.selectAvailable(ctx)
	if err != nil {
		return err
	}
	return p.Speak(ctx, text)
}

func (a *autoProvider) Synthesize(ctx context.Context, text string) ([]byte, error) {
	p, err := a.selectAvailable(ctx)
	if err != nil {
		return nil, err
	}
	return p.Synthesize(ctx, text)
}

func (a *autoProvider) SetVoice(voice string) error {
	var lastErr error
	for _, p := range a.providers {
		if err := p.SetVoice(voice); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (a *autoProvider) SetSpeed(speed float64) {
	for _, p := range a.providers {
		p.SetSpeed(speed)
	}
}

func (a *autoProvider) Stop() {
	for _, p := range a.providers {
		p.Stop()
	}
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
		return fmt.Errorf("errors closing TTS providers: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (a *autoProvider) selectAvailable(ctx context.Context) (Provider, error) {
	for _, p := range a.providers {
		if p.IsAvailable(ctx) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("no TTS provider available")
}
