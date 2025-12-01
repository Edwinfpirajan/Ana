package stt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/anastreamer/ana/pkg/utils"
	"github.com/rs/zerolog"
)

// WhisperProvider implements STT using local whisper.cpp
type WhisperProvider struct {
	binaryPath string
	modelPath  string
	language   string
	log        zerolog.Logger
}

// NewWhisperProvider creates a new Whisper provider
func NewWhisperProvider(cfg config.WhisperConfig) (*WhisperProvider, error) {
	binaryPath := utils.GetBinaryPath(cfg.BinaryPath)

	// Convert model path to absolute path
	modelPath, err := filepath.Abs(cfg.ModelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve model path %s: %w", cfg.ModelPath, err)
	}

	return &WhisperProvider{
		binaryPath: binaryPath,
		modelPath:  modelPath,
		language:   cfg.Language,
		log:        logger.Component("whisper"),
	}, nil
}

// Name returns the provider name
func (p *WhisperProvider) Name() string {
	return "whisper"
}

// Transcribe converts audio bytes to text
func (p *WhisperProvider) Transcribe(ctx context.Context, audio []byte) (*TranscriptionResult, error) {
	p.log.Debug().
		Int("audio_len", len(audio)).
		Msg("Transcribe called with audio bytes")

	// Create temporary WAV file
	tempFile := utils.GetTempFilePath("ana_audio", ".wav")
	defer os.Remove(tempFile)

	// Check if audio is already WAV format or raw PCM
	if len(audio) > 4 && string(audio[0:4]) == "RIFF" {
		// Already WAV format
		if err := os.WriteFile(tempFile, audio, 0644); err != nil {
			return nil, fmt.Errorf("failed to write temp audio file: %w", err)
		}
		p.log.Debug().Str("file", tempFile).Int("size", len(audio)).Msg("Wrote WAV file (already WAV format)")
	} else {
		// Raw PCM, convert to WAV
		if err := utils.SaveWAV(tempFile, audio, 16000, 1, 16); err != nil {
			return nil, fmt.Errorf("failed to save audio as WAV: %w", err)
		}
		p.log.Debug().Str("file", tempFile).Int("pcm_size", len(audio)).Msg("Converted PCM to WAV")
	}

	// Verify file was created
	fileInfo, err := os.Stat(tempFile)
	if err != nil {
		p.log.Error().Err(err).Str("file", tempFile).Msg("Failed to stat WAV file")
		return nil, fmt.Errorf("failed to verify WAV file: %w", err)
	}
	p.log.Debug().Str("file", tempFile).Int64("size", fileInfo.Size()).Msg("WAV file created successfully")

	return p.TranscribeFile(ctx, tempFile)
}

// TranscribeFile transcribes an audio file
func (p *WhisperProvider) TranscribeFile(ctx context.Context, filePath string) (*TranscriptionResult, error) {
	start := time.Now()

	// Get file size for debugging
	fileInfo, _ := os.Stat(filePath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	p.log.Debug().
		Str("file", filePath).
		Int64("file_size", fileSize).
		Str("model", p.modelPath).
		Str("language", p.language).
		Msg("Starting transcription")

	// Build command arguments for whisper-cli
	args := []string{
		"-m", p.modelPath,           // model path
		"-l", p.language,            // language
		"-otxt",                     // output as text file
		"-of", filePath[:len(filePath)-4], // output file (without extension)
		"-nt",                       // no timestamps
		"-f", filePath,              // input file
	}

	// Run whisper
	result, err := utils.RunProcess(ctx, p.binaryPath, args...)
	if err != nil {
		return nil, fmt.Errorf("whisper execution failed: %w", err)
	}

	p.log.Debug().
		Int("exit_code", result.ExitCode).
		Int("stdout_len", len(result.Stdout)).
		Int("stderr_len", len(result.Stderr)).
		Msg("Whisper process completed")

	if result.ExitCode != 0 {
		p.log.Error().
			Int("exit_code", result.ExitCode).
			Str("stderr", result.Stderr).
			Msg("Whisper failed")
		return nil, fmt.Errorf("whisper returned non-zero exit code: %d, stderr: %s", result.ExitCode, result.Stderr)
	}

	// Parse output - whisper outputs to stdout
	text := strings.TrimSpace(result.Stdout)
	p.log.Debug().
		Str("stdout_content", result.Stdout).
		Str("trimmed_text", text).
		Msg("Whisper stdout content")

	// Also check for output file specified via -of flag
	outputPath := filePath[:len(filePath)-4] // Remove .wav extension
	txtFile := outputPath + ".txt"
	p.log.Debug().Str("expected_txt_file", txtFile).Msg("Looking for output file")

	if _, err := os.Stat(txtFile); err == nil {
		content, err := os.ReadFile(txtFile)
		if err == nil {
			fileText := strings.TrimSpace(string(content))
			p.log.Debug().
				Int("file_size", len(content)).
				Str("file_content", fileText).
				Msg("Read text from output file")
			text = fileText
		} else {
			p.log.Error().Err(err).Str("file", txtFile).Msg("Failed to read output file")
		}
		os.Remove(txtFile) // Clean up
	} else {
		p.log.Debug().Err(err).Str("file", txtFile).Msg("Output file not found")
	}

	// Clean up the text
	text = cleanTranscription(text)

	p.log.Debug().
		Str("text", text).
		Dur("duration", time.Since(start)).
		Msg("Transcription complete")

	return &TranscriptionResult{
		Text:       text,
		Language:   p.language,
		Confidence: 1.0, // Whisper doesn't provide confidence
		Duration:   result.Duration.Seconds(),
	}, nil
}

// SetLanguage sets the language for transcription
func (p *WhisperProvider) SetLanguage(lang string) {
	p.language = lang
}

// IsAvailable checks if Whisper is available
func (p *WhisperProvider) IsAvailable(ctx context.Context) bool {
	// Check binary exists
	if !utils.BinaryExists(p.binaryPath) {
		p.log.Warn().Str("path", p.binaryPath).Msg("Whisper binary not found")
		return false
	}

	// Check model exists
	if !utils.FileExists(p.modelPath) {
		p.log.Warn().Str("path", p.modelPath).Msg("Whisper model not found")
		return false
	}

	return true
}

// Close releases resources
func (p *WhisperProvider) Close() error {
	return nil
}

// cleanTranscription cleans up whisper output
func cleanTranscription(text string) string {
	// Remove common whisper artifacts
	text = strings.TrimSpace(text)

	// Remove [BLANK_AUDIO] markers
	text = strings.ReplaceAll(text, "[BLANK_AUDIO]", "")

	// Remove timing markers like [00:00:00.000 --> 00:00:02.000]
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip timestamp lines
		if strings.HasPrefix(line, "[") && strings.Contains(line, "-->") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	return strings.Join(cleanLines, " ")
}
