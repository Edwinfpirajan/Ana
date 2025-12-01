package tts

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/anastreamer/ana/pkg/utils"
	"github.com/rs/zerolog"
)

// PiperProvider implements TTS using local Piper
type PiperProvider struct {
	binaryPath string
	modelPath  string
	speed      float64
	log        zerolog.Logger

	mu         sync.Mutex
	currentCmd *exec.Cmd
	isPlaying  bool
}

// NewPiperProvider creates a new Piper TTS provider
func NewPiperProvider(cfg config.PiperConfig) (*PiperProvider, error) {
	binaryPath := utils.GetBinaryPath(cfg.BinaryPath)

	// Expand paths to absolute
	if !filepath.IsAbs(binaryPath) {
		absPath, err := filepath.Abs(binaryPath)
		if err == nil {
			binaryPath = absPath
		}
	}

	modelPath := cfg.ModelPath
	if !filepath.IsAbs(modelPath) {
		absPath, err := filepath.Abs(modelPath)
		if err == nil {
			modelPath = absPath
		}
	}

	speed := cfg.Speed
	if speed == 0 {
		speed = 1.0
	}

	return &PiperProvider{
		binaryPath: binaryPath,
		modelPath:  modelPath,
		speed:      speed,
		log:        logger.Component("piper"),
	}, nil
}

// Name returns the provider name
func (p *PiperProvider) Name() string {
	return "piper"
}

// Speak converts text to speech and plays it
func (p *PiperProvider) Speak(ctx context.Context, text string) error {
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

	p.log.Debug().Str("text", text).Msg("Speaking text")

	// Synthesize audio
	audio, err := p.Synthesize(ctx, text)
	if err != nil {
		return err
	}

	// Play the audio
	return p.playAudio(ctx, audio)
}

// Synthesize converts text to audio bytes
func (p *PiperProvider) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}

	p.log.Debug().
		Str("binary", p.binaryPath).
		Str("model", p.modelPath).
		Msg("Piper synthesize starting")

	// Create temp output file
	tempFile := utils.GetTempFilePath("ana_tts", ".wav")
	defer os.Remove(tempFile)

	// Ensure temp directory exists and is writable
	tempDir := filepath.Dir(tempFile)
	if err := utils.EnsureDir(tempDir); err != nil {
		return nil, fmt.Errorf("failed to ensure temp directory exists: %w", err)
	}

	// Build command arguments
	args := []string{
		"--model", p.modelPath,
		"--output_file", tempFile,
	}

	// Add speed/length scale if not default
	if p.speed != 1.0 {
		args = append(args, "--length-scale", fmt.Sprintf("%.2f", 1.0/p.speed))
	}

	p.log.Debug().
		Strs("args", args).
		Msg("Piper command arguments")

	// Create command
	cmd := exec.CommandContext(ctx, p.binaryPath, args...)
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	// Pipe text to stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		p.log.Error().
			Err(err).
			Str("binary", p.binaryPath).
			Msg("Failed to start piper")
		return nil, fmt.Errorf("failed to start piper: %w", err)
	}

	// Write text to stdin
	if _, err := stdin.Write([]byte(text)); err != nil {
		return nil, fmt.Errorf("failed to write to stdin: %w", err)
	}
	stdin.Close()

	// Wait for completion
	if err := cmd.Wait(); err != nil {
		errMsg := fmt.Sprintf("piper execution failed: %v", err)
		if stderr.Len() > 0 {
			errMsg = fmt.Sprintf("%s (stderr: %s)", errMsg, stderr.String())
		}
		if stdout.Len() > 0 {
			errMsg = fmt.Sprintf("%s (stdout: %s)", errMsg, stdout.String())
		}
		p.log.Error().Str("error", errMsg).Msg("Piper failed")
		return nil, fmt.Errorf(errMsg)
	}

	// Read output file
	audio, err := os.ReadFile(tempFile)
	if err != nil {
		p.log.Error().
			Err(err).
			Str("tempFile", tempFile).
			Msg("Failed to read piper output")
		return nil, fmt.Errorf("failed to read output audio: %w", err)
	}

	p.log.Debug().Int("audioSize", len(audio)).Msg("Piper synthesis complete")
	return audio, nil
}

// playAudio plays WAV audio data
func (p *PiperProvider) playAudio(ctx context.Context, audio []byte) error {
	// Create temp file for playback
	tempFile := utils.GetTempFilePath("ana_play", ".wav")
	defer os.Remove(tempFile)

	// Ensure temp directory exists
	tempDir := filepath.Dir(tempFile)
	if err := utils.EnsureDir(tempDir); err != nil {
		return fmt.Errorf("failed to ensure temp directory exists: %w", err)
	}

	if err := os.WriteFile(tempFile, audio, 0644); err != nil {
		return fmt.Errorf("failed to write temp audio: %w", err)
	}

	// Verify file was created successfully
	if !utils.FileExists(tempFile) {
		return fmt.Errorf("failed to create temp audio file: %s", tempFile)
	}

	// Use platform-specific player
	var cmd *exec.Cmd

	// Try different audio players
	players := []struct {
		name string
		args []string
	}{
		{"aplay", []string{"-q", tempFile}}, // Linux ALSA
		{"paplay", []string{tempFile}},      // Linux PulseAudio
		{"afplay", []string{tempFile}},      // macOS
		{"powershell", []string{"-c", fmt.Sprintf(`(New-Object Media.SoundPlayer '%s').PlaySync()`, tempFile)}}, // Windows
	}

	for _, player := range players {
		if _, err := exec.LookPath(player.name); err == nil {
			cmd = exec.CommandContext(ctx, player.name, player.args...)
			break
		}
	}

	if cmd == nil {
		return fmt.Errorf("no audio player found")
	}

	p.mu.Lock()
	p.currentCmd = cmd
	p.mu.Unlock()

	if err := cmd.Run(); err != nil {
		// Check if it was cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("failed to play audio: %w", err)
	}

	return nil
}

// SetVoice sets the voice model to use
func (p *PiperProvider) SetVoice(voice string) error {
	if !utils.FileExists(voice) {
		return fmt.Errorf("voice model not found: %s", voice)
	}
	p.modelPath = voice
	return nil
}

// SetSpeed sets the speech speed
func (p *PiperProvider) SetSpeed(speed float64) {
	if speed <= 0 {
		speed = 1.0
	}
	p.speed = speed
}

// Stop stops any current playback
func (p *PiperProvider) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentCmd != nil && p.currentCmd.Process != nil {
		p.currentCmd.Process.Kill()
	}
	p.isPlaying = false
}

// IsAvailable checks if Piper is available
func (p *PiperProvider) IsAvailable(ctx context.Context) bool {
	// Check binary exists
	if !utils.BinaryExists(p.binaryPath) {
		p.log.Warn().Str("path", p.binaryPath).Msg("Piper binary not found")
		return false
	}

	// Check model exists
	if !utils.FileExists(p.modelPath) {
		p.log.Warn().Str("path", p.modelPath).Msg("Piper model not found")
		return false
	}

	return true
}

// Close releases resources
func (p *PiperProvider) Close() error {
	p.Stop()
	return nil
}
