package tts

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/rs/zerolog"
)

// WindowsTTSProvider implements TTS using Windows built-in SAPI
type WindowsTTSProvider struct {
	speed      float64
	voice      string
	log        zerolog.Logger
	mu         sync.Mutex
	isPlaying  bool
}

// NewWindowsTTSProvider creates a new Windows TTS provider
func NewWindowsTTSProvider(cfg config.OpenAITTSConfig) (*WindowsTTSProvider, error) {
	voice := cfg.Voice
	if voice == "" {
		voice = "nova"
	}

	return &WindowsTTSProvider{
		speed: 1.0,
		voice: voice,
		log:   logger.Component("windows-tts"),
	}, nil
}

// Name returns the provider name
func (w *WindowsTTSProvider) Name() string {
	return "windows-tts"
}

// Synthesize converts text to audio bytes (not implemented for Windows TTS)
func (w *WindowsTTSProvider) Synthesize(ctx context.Context, text string) ([]byte, error) {
	// Windows TTS doesn't support synthesis to bytes, only speaking
	return nil, fmt.Errorf("synthesize not supported for windows-tts, use Speak instead")
}

// Speak converts text to speech using Windows SAPI and plays it
func (w *WindowsTTSProvider) Speak(ctx context.Context, text string) error {
	w.mu.Lock()
	if w.isPlaying {
		w.mu.Unlock()
		return nil
	}
	w.isPlaying = true
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.isPlaying = false
		w.mu.Unlock()
	}()

	w.log.Debug().Str("text", text).Msg("Speaking text with Windows TTS")

	// Create PowerShell script to speak with Spanish voice
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Speech
$speak = New-Object System.Speech.Synthesis.SpeechSynthesizer
try {
    $voice = $speak.GetInstalledVoices() | Where-Object {$_.VoiceInfo.Culture.Name -match "es"} | Select-Object -First 1
    if ($voice) {
        $speak.SelectVoice($voice.VoiceInfo.Name)
    }
}
catch {}
$speak.Rate = %d
$speak.Speak("%s")
`,
		int(w.speed*10)-10, // Rate from -10 to 10, default is 0
		escapeQuotes(text),
	)

	// Execute PowerShell
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", script)

	if err := cmd.Run(); err != nil {
		w.log.Warn().Err(err).Msg("Windows TTS failed, but continuing")
		return nil // Don't fail, just warn
	}

	return nil
}

// IsAvailable checks if Windows TTS is available
func (w *WindowsTTSProvider) IsAvailable(ctx context.Context) bool {
	// Windows TTS is always available on Windows
	return true
}

// Close releases resources
func (w *WindowsTTSProvider) Close() error {
	return nil
}

// Stop stops any current playback (not implemented for Windows TTS)
func (w *WindowsTTSProvider) Stop() {
	// Windows SAPI handles this internally
}

// SetVoice sets the voice (not fully implemented for Windows TTS)
func (w *WindowsTTSProvider) SetVoice(voice string) error {
	w.voice = voice
	return nil
}

// SetSpeed sets the speech speed
func (w *WindowsTTSProvider) SetSpeed(speed float64) {
	if speed <= 0 {
		speed = 1.0
	}
	w.speed = speed
}

// escapeQuotes escapes quotes in the text for PowerShell
func escapeQuotes(text string) string {
	return strings.ReplaceAll(text, `"`, `\"`)
}
