package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/anastreamer/ana/pkg/utils"
	"github.com/rs/zerolog"
)

const openAITTSURL = "https://api.openai.com/v1/audio/speech"

// OpenAITTSProvider implements TTS using OpenAI's API
type OpenAITTSProvider struct {
	apiKey string
	model  string
	voice  string
	speed  float64
	client *http.Client
	log    zerolog.Logger

	mu         sync.Mutex
	currentCmd *exec.Cmd
	isPlaying  bool
}

// OpenAITTSRequest represents the API request
type OpenAITTSRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

// NewOpenAIProvider creates a new OpenAI TTS provider
func NewOpenAIProvider(cfg config.OpenAITTSConfig) (*OpenAITTSProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	model := cfg.Model
	if model == "" {
		model = "tts-1"
	}

	voice := cfg.Voice
	if voice == "" {
		voice = "nova"
	}

	return &OpenAITTSProvider{
		apiKey: cfg.APIKey,
		model:  model,
		voice:  voice,
		speed:  1.0,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		log: logger.Component("openai-tts"),
	}, nil
}

// Name returns the provider name
func (p *OpenAITTSProvider) Name() string {
	return "openai-tts"
}

// Speak converts text to speech and plays it
func (p *OpenAITTSProvider) Speak(ctx context.Context, text string) error {
	p.mu.Lock()
	if p.isPlaying {
		p.mu.Unlock()
		p.Stop()
		p.mu.Lock()
	}
	p.isPlaying = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.isPlaying = false
		p.currentCmd = nil
		p.mu.Unlock()
	}()

	p.log.Debug().Str("text", text).Msg("Speaking text via OpenAI")

	// Synthesize audio
	audio, err := p.Synthesize(ctx, text)
	if err != nil {
		return err
	}

	// Play the audio
	return p.playAudio(ctx, audio)
}

// Synthesize converts text to audio bytes (MP3 format from OpenAI)
func (p *OpenAITTSProvider) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}

	// Create request
	reqBody := OpenAITTSRequest{
		Model:          p.model,
		Input:          text,
		Voice:          p.voice,
		ResponseFormat: "mp3", // OpenAI returns MP3 by default
		Speed:          p.speed,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", openAITTSURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read audio data
	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return audio, nil
}

// playAudio plays audio data (MP3 format)
func (p *OpenAITTSProvider) playAudio(ctx context.Context, audio []byte) error {
	// Create temp file for playback
	tempFile := utils.GetTempFilePath("ana_play", ".mp3")
	defer os.Remove(tempFile)

	if err := os.WriteFile(tempFile, audio, 0644); err != nil {
		return fmt.Errorf("failed to write temp audio: %w", err)
	}

	// Use platform-specific player
	var cmd *exec.Cmd

	// Try different audio players
	players := []struct {
		name string
		args []string
	}{
		{"mpv", []string{"--no-video", "--really-quiet", tempFile}},                  // Cross-platform
		{"ffplay", []string{"-nodisp", "-autoexit", "-loglevel", "quiet", tempFile}}, // FFmpeg
		{"afplay", []string{tempFile}},                                               // macOS
		{"powershell", []string{"-c", fmt.Sprintf(`Add-Type -AssemblyName presentationCore; $player = New-Object system.windows.media.mediaplayer; $player.open('%s'); Start-Sleep 1; while($player.Position -lt $player.NaturalDuration.TimeSpan) { Start-Sleep -Milliseconds 100 }`, tempFile)}}, // Windows
	}

	for _, player := range players {
		if _, err := exec.LookPath(player.name); err == nil {
			cmd = exec.CommandContext(ctx, player.name, player.args...)
			break
		}
	}

	if cmd == nil {
		return fmt.Errorf("no MP3 audio player found (install mpv or ffplay)")
	}

	p.mu.Lock()
	p.currentCmd = cmd
	p.mu.Unlock()

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("failed to play audio: %w", err)
	}

	return nil
}

// SetVoice sets the voice to use
func (p *OpenAITTSProvider) SetVoice(voice string) error {
	validVoices := []string{"alloy", "echo", "fable", "onyx", "nova", "shimmer"}
	for _, v := range validVoices {
		if v == voice {
			p.voice = voice
			return nil
		}
	}
	return fmt.Errorf("invalid voice: %s (valid: alloy, echo, fable, onyx, nova, shimmer)", voice)
}

// SetSpeed sets the speech speed
func (p *OpenAITTSProvider) SetSpeed(speed float64) {
	if speed < 0.25 {
		speed = 0.25
	}
	if speed > 4.0 {
		speed = 4.0
	}
	p.speed = speed
}

// Stop stops any current playback
func (p *OpenAITTSProvider) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentCmd != nil && p.currentCmd.Process != nil {
		p.currentCmd.Process.Kill()
	}
	p.isPlaying = false
}

// IsAvailable checks if OpenAI TTS is available
func (p *OpenAITTSProvider) IsAvailable(ctx context.Context) bool {
	return p.apiKey != ""
}

// Close releases resources
func (p *OpenAITTSProvider) Close() error {
	p.Stop()
	return nil
}
