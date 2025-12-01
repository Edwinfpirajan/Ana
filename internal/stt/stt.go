// Package stt provides Speech-to-Text functionality for AnaStreamer
package stt

import (
	"context"
	"fmt"

	"github.com/anastreamer/ana/internal/config"
)

// TranscriptionResult holds the result of a transcription
type TranscriptionResult struct {
	Text       string  // Transcribed text
	Language   string  // Detected language
	Confidence float64 // Confidence score (0-1)
	Duration   float64 // Audio duration in seconds
}

// Provider is the interface for STT providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Transcribe converts audio bytes to text
	// Audio should be in WAV format, 16kHz, mono, 16-bit PCM
	Transcribe(ctx context.Context, audio []byte) (*TranscriptionResult, error)

	// TranscribeFile transcribes an audio file
	TranscribeFile(ctx context.Context, filePath string) (*TranscriptionResult, error)

	// SetLanguage sets the language for transcription
	SetLanguage(lang string)

	// IsAvailable checks if the provider is available
	IsAvailable(ctx context.Context) bool

	// Close releases any resources
	Close() error
}

// New creates a new STT provider based on configuration
func New(cfg *config.Config) (Provider, error) {
	switch cfg.STT.Provider {
	case "whisper":
		return NewWhisperProvider(cfg.STT.Whisper)
	case "openai":
		return NewOpenAIProvider(cfg.STT.OpenAI)
	default:
		return nil, fmt.Errorf("unknown STT provider: %s", cfg.STT.Provider)
	}
}
