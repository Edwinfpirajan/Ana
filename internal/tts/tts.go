// Package tts provides Text-to-Speech functionality for AnaStreamer
package tts

import (
	"context"
	"fmt"

	"github.com/anastreamer/ana/internal/config"
)

// Provider is the interface for TTS providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Speak converts text to speech and plays it
	Speak(ctx context.Context, text string) error

	// Synthesize converts text to audio bytes (WAV format)
	Synthesize(ctx context.Context, text string) ([]byte, error)

	// SetVoice sets the voice to use
	SetVoice(voice string) error

	// SetSpeed sets the speech speed (1.0 = normal)
	SetSpeed(speed float64)

	// Stop stops any current playback
	Stop()

	// IsAvailable checks if the provider is available
	IsAvailable(ctx context.Context) bool

	// Close releases any resources
	Close() error
}

// New creates a new TTS provider based on configuration
func New(cfg *config.Config) (Provider, error) {
	switch cfg.TTS.Provider {
	case "piper":
		return NewPiperProvider(cfg.TTS.Piper)
	case "openai":
		return NewOpenAIProvider(cfg.TTS.OpenAI)
	case "auto":
		return NewAutoProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown TTS provider: %s", cfg.TTS.Provider)
	}
}
