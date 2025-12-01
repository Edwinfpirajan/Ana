package stt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/anastreamer/ana/pkg/utils"
	"github.com/rs/zerolog"
)

const openAISTTURL = "https://api.openai.com/v1/audio/transcriptions"

// OpenAISTTProvider implements STT using OpenAI's Whisper API
type OpenAISTTProvider struct {
	apiKey   string
	model    string
	language string
	client   *http.Client
	log      zerolog.Logger
}

// OpenAITranscriptionResponse represents the API response
type OpenAITranscriptionResponse struct {
	Text     string  `json:"text"`
	Language string  `json:"language,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

// NewOpenAIProvider creates a new OpenAI STT provider
func NewOpenAIProvider(cfg config.OpenAISTTConfig) (*OpenAISTTProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	model := cfg.Model
	if model == "" {
		model = "whisper-1"
	}

	return &OpenAISTTProvider{
		apiKey:   cfg.APIKey,
		model:    model,
		language: "es", // Default language
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		log: logger.Component("openai-stt"),
	}, nil
}

// Name returns the provider name
func (p *OpenAISTTProvider) Name() string {
	return "openai-stt"
}

// Transcribe converts audio bytes to text
func (p *OpenAISTTProvider) Transcribe(ctx context.Context, audio []byte) (*TranscriptionResult, error) {
	// Validate audio input
	if audio == nil || len(audio) == 0 {
		return nil, fmt.Errorf("audio data is empty or nil")
	}

	// Create temporary WAV file
	tempFile := utils.GetTempFilePath("ana_audio", ".wav")
	defer os.Remove(tempFile)

	// Ensure temp directory exists
	tempDir := os.TempDir()
	if err := utils.EnsureDir(tempDir); err != nil {
		return nil, fmt.Errorf("failed to ensure temp directory: %w", err)
	}

	// Check if audio is already WAV format or raw PCM
	if len(audio) > 4 && string(audio[0:4]) == "RIFF" {
		if err := os.WriteFile(tempFile, audio, 0644); err != nil {
			return nil, fmt.Errorf("failed to write temp audio file: %w", err)
		}
	} else {
		if err := utils.SaveWAV(tempFile, audio, 16000, 1, 16); err != nil {
			return nil, fmt.Errorf("failed to save audio as WAV: %w", err)
		}
	}

	// Verify temp file was created
	if !utils.FileExists(tempFile) {
		return nil, fmt.Errorf("failed to create temporary audio file")
	}

	return p.TranscribeFile(ctx, tempFile)
}

// TranscribeFile transcribes an audio file
func (p *OpenAISTTProvider) TranscribeFile(ctx context.Context, filePath string) (*TranscriptionResult, error) {
	start := time.Now()

	p.log.Debug().
		Str("file", filePath).
		Str("language", p.language).
		Msg("Starting OpenAI transcription")

	// Open audio file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy audio data: %w", err)
	}

	// Add model
	if err := writer.WriteField("model", p.model); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	// Add language
	if p.language != "" {
		if err := writer.WriteField("language", p.language); err != nil {
			return nil, fmt.Errorf("failed to write language field: %w", err)
		}
	}

	// Add response format
	if err := writer.WriteField("response_format", "json"); err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", openAISTTURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var transcriptionResp OpenAITranscriptionResponse
	if err := json.Unmarshal(body, &transcriptionResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	p.log.Debug().
		Str("text", transcriptionResp.Text).
		Dur("duration", time.Since(start)).
		Msg("Transcription complete")

	return &TranscriptionResult{
		Text:       transcriptionResp.Text,
		Language:   transcriptionResp.Language,
		Confidence: 1.0, // OpenAI doesn't provide confidence
		Duration:   time.Since(start).Seconds(),
	}, nil
}

// SetLanguage sets the language for transcription
func (p *OpenAISTTProvider) SetLanguage(lang string) {
	p.language = lang
}

// IsAvailable checks if OpenAI STT is available
func (p *OpenAISTTProvider) IsAvailable(ctx context.Context) bool {
	// Simple check - verify API key is set
	return p.apiKey != ""
}

// Close releases resources
func (p *OpenAISTTProvider) Close() error {
	return nil
}
