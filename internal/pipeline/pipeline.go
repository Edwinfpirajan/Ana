// Package pipeline provides the main processing pipeline for AnaStreamer
package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anastreamer/ana/internal/brain"
	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/internal/llm"
	"github.com/anastreamer/ana/internal/stt"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/anastreamer/ana/pkg/utils"
	"github.com/rs/zerolog"
)

// State represents the current state of the pipeline
type State int

const (
	StateIdle State = iota
	StateListening
	StateRecording
	StateProcessing
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateListening:
		return "listening"
	case StateRecording:
		return "recording"
	case StateProcessing:
		return "processing"
	default:
		return "unknown"
	}
}

// Pipeline is the main audio processing pipeline
type Pipeline struct {
	cfg         *config.Config
	sttProvider stt.Provider
	brain       *brain.Brain
	log         zerolog.Logger

	state   State
	stateMu sync.RWMutex

	// Audio buffer for recording
	audioBuffer *bytes.Buffer
	bufferMu    sync.Mutex

	// Channels for events
	wakeWordChan chan struct{}
	hotkeyDown   chan struct{}
	hotkeyUp     chan struct{}
	audioChan    chan []byte
	stopChan     chan struct{}

	// VAD settings
	silenceStart time.Time
	speechStart  time.Time
	hasSpeech    bool

	// Callbacks
	onStateChange func(State)
	onTranscript  func(string)
	onResponse    func(string)
	onError       func(error)
}

// NewPipeline creates a new processing pipeline
func NewPipeline(cfg *config.Config, sttProvider stt.Provider, brn *brain.Brain) *Pipeline {
	return &Pipeline{
		cfg:          cfg,
		sttProvider:  sttProvider,
		brain:        brn,
		log:          logger.Component("pipeline"),
		state:        StateIdle,
		audioBuffer:  bytes.NewBuffer(nil),
		wakeWordChan: make(chan struct{}, 1),
		hotkeyDown:   make(chan struct{}, 1),
		hotkeyUp:     make(chan struct{}, 1),
		audioChan:    make(chan []byte, 100),
		stopChan:     make(chan struct{}),
	}
}

// SetCallbacks sets the callback functions
func (p *Pipeline) SetCallbacks(
	onStateChange func(State),
	onTranscript func(string),
	onResponse func(string),
	onError func(error),
) {
	p.onStateChange = onStateChange
	p.onTranscript = onTranscript
	p.onResponse = onResponse
	p.onError = onError
}

// setState updates the pipeline state
func (p *Pipeline) setState(state State) {
	p.stateMu.Lock()
	oldState := p.state
	p.state = state
	p.stateMu.Unlock()

	if oldState != state {
		p.log.Debug().
			Str("from", oldState.String()).
			Str("to", state.String()).
			Msg("State change")

		if p.onStateChange != nil {
			p.onStateChange(state)
		}
	}
}

// GetState returns the current state
func (p *Pipeline) GetState() State {
	p.stateMu.RLock()
	defer p.stateMu.RUnlock()
	return p.state
}

// Start starts the pipeline
func (p *Pipeline) Start(ctx context.Context) error {
	p.log.Info().Msg("Starting pipeline")
	p.setState(StateIdle)

	go p.run(ctx)
	return nil
}

// Stop stops the pipeline
func (p *Pipeline) Stop() {
	p.log.Info().Msg("Stopping pipeline")
	close(p.stopChan)
}

// TriggerWakeWord simulates a wake word detection
func (p *Pipeline) TriggerWakeWord() {
	select {
	case p.wakeWordChan <- struct{}{}:
	default:
	}
}

// TriggerHotkeyDown simulates hotkey press
func (p *Pipeline) TriggerHotkeyDown() {
	select {
	case p.hotkeyDown <- struct{}{}:
	default:
	}
}

// TriggerHotkeyUp simulates hotkey release
func (p *Pipeline) TriggerHotkeyUp() {
	select {
	case p.hotkeyUp <- struct{}{}:
	default:
	}
}

// FeedAudio feeds audio data to the pipeline
func (p *Pipeline) FeedAudio(audio []byte) {
	select {
	case p.audioChan <- audio:
	default:
		// Drop if buffer is full
	}
}

// ProcessText directly processes a text command (for testing)
func (p *Pipeline) ProcessText(ctx context.Context, text string) (string, error) {
	p.setState(StateProcessing)
	defer p.setState(StateIdle)

	// Check if Ana name is mentioned
	if !llm.IsAnaActivated(text) {
		p.log.Debug().Str("text", text).Msg("Ignoring input - Ana name not mentioned")
		return "", nil
	}

	if p.onTranscript != nil {
		p.onTranscript(text)
	}

	response, err := p.brain.ProcessCommand(ctx, text)
	if err == nil && response != "" {
		if p.onResponse != nil {
			p.onResponse(response)
		}
		// Try to speak but don't fail if it doesn't work
		_ = p.brain.ProcessAndSpeak(ctx, text)
	}

	return response, err
}

// run is the main pipeline loop
func (p *Pipeline) run(ctx context.Context) {
	p.log.Debug().Msg("Pipeline loop started")

	for {
		select {
		case <-ctx.Done():
			p.log.Debug().Msg("Pipeline context done")
			return

		case <-p.stopChan:
			p.log.Debug().Msg("Pipeline stop signal")
			return

		case <-p.wakeWordChan:
			p.log.Info().Msg("Wake word detected")
			p.handleWakeWord(ctx)

		case <-p.hotkeyDown:
			p.log.Debug().Msg("Hotkey pressed")
			p.startRecording()

		case <-p.hotkeyUp:
			p.log.Debug().Msg("Hotkey released")
			p.handleHotkeyRelease(ctx)

		case audio := <-p.audioChan:
			p.handleAudio(ctx, audio)
		}
	}
}

// handleWakeWord handles wake word detection
func (p *Pipeline) handleWakeWord(ctx context.Context) {
	p.setState(StateListening)
	p.log.Info().Msg("Ana activated - entering persistent session mode")
	if p.onResponse != nil {
		p.onResponse("Aquí estoy! Dime en qué puedo ayudarte.")
	}

	// Start persistent listening session
	// The session continues until user says deactivation word
	go p.persistentListeningSession(ctx)
}

// persistentListeningSession handles continuous listening after wake word detection
func (p *Pipeline) persistentListeningSession(ctx context.Context) {
	// Keep listening until deactivation or context cancellation
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		default:
		}

		// Check if still in listening state
		if p.GetState() != StateListening {
			return
		}

		// Record until silence
		p.startRecording()
		p.recordUntilSilence(ctx)
		p.processRecordedAudio(ctx)

		// processRecordedAudio will set state to StateIdle if deactivation detected
		// Otherwise it will keep us in StateListening for next command
	}
}

// handleHotkeyRelease handles when the hotkey is released
func (p *Pipeline) handleHotkeyRelease(ctx context.Context) {
	if p.GetState() != StateRecording {
		return
	}

	p.processRecordedAudio(ctx)
}

// startRecording starts recording audio
func (p *Pipeline) startRecording() {
	p.bufferMu.Lock()
	p.audioBuffer.Reset()
	p.bufferMu.Unlock()

	p.setState(StateRecording)
	p.hasSpeech = false
	p.silenceStart = time.Time{}
	p.speechStart = time.Time{}
}

// handleAudio handles incoming audio chunks
func (p *Pipeline) handleAudio(ctx context.Context, audio []byte) {
	state := p.GetState()

	// Auto-start recording when speech is detected in Idle state
	if state == StateIdle && p.cfg.Audio.VAD.Enabled {
		samples := utils.BytesToInt16(audio)
		rms := utils.CalculateRMS(samples)
		threshold := float64(500) * (1.0 - p.cfg.Audio.VAD.Sensitivity)

		if rms > threshold {
			p.log.Debug().Float64("rms", rms).Float64("threshold", threshold).Msg("Voice detected in Idle state, starting recording")
			p.setState(StateListening)
			p.startRecording()

			// Start silence detection in background
			go func() {
				p.recordUntilSilence(ctx)
				p.processRecordedAudio(ctx)
			}()
			// Fall through to record this chunk
		}
	}

	// Always analyze for VAD when listening or recording
	if state == StateListening || state == StateRecording {
		p.bufferMu.Lock()
		p.audioBuffer.Write(audio)
		p.bufferMu.Unlock()

		// Check for speech/silence
		if p.cfg.Audio.VAD.Enabled {
			p.analyzeVAD(audio)
		}
	}
}

// analyzeVAD analyzes audio for Voice Activity Detection
func (p *Pipeline) analyzeVAD(audio []byte) {
	// Convert bytes to samples
	samples := utils.BytesToInt16(audio)

	// Calculate RMS energy
	rms := utils.CalculateRMS(samples)

	// Threshold based on sensitivity
	// Higher sensitivity = lower threshold = more sensitive to quiet sounds
	threshold := float64(500) * (1.0 - p.cfg.Audio.VAD.Sensitivity)

	isSpeech := rms > threshold

	now := time.Now()

	if isSpeech {
		if !p.hasSpeech {
			p.speechStart = now
			p.hasSpeech = true
			p.log.Debug().Float64("rms", rms).Msg("Speech detected")
		}
		p.silenceStart = time.Time{} // Reset silence timer
	} else {
		if p.hasSpeech && p.silenceStart.IsZero() {
			p.silenceStart = now
			p.log.Debug().Float64("rms", rms).Msg("Silence detected")
		}
	}
}

// recordUntilSilence records until silence is detected
func (p *Pipeline) recordUntilSilence(ctx context.Context) {
	silenceThreshold := time.Duration(p.cfg.Audio.VAD.SilenceThresholdMs) * time.Millisecond
	maxRecordTime := 30 * time.Second // Maximum recording time

	timeout := time.NewTimer(maxRecordTime)
	defer timeout.Stop()

	checkInterval := 100 * time.Millisecond
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case <-timeout.C:
			p.log.Warn().Msg("Recording timeout reached")
			return
		case <-ticker.C:
			// Check if we have enough silence after speech
			if p.hasSpeech && !p.silenceStart.IsZero() {
				silenceDuration := time.Since(p.silenceStart)
				if silenceDuration >= silenceThreshold {
					p.log.Debug().
						Dur("silence", silenceDuration).
						Msg("Silence threshold reached")
					return
				}
			}
		}
	}
}

// processRecordedAudio processes the recorded audio buffer
func (p *Pipeline) processRecordedAudio(ctx context.Context) {
	p.setState(StateProcessing)
	// Don't immediately return to Idle - check for deactivation first
	defer func() {
		// Only return to Idle if not in a persistent session or if deactivated
		if p.GetState() == StateProcessing {
			p.setState(StateIdle)
		}
	}()

	p.bufferMu.Lock()
	audio := p.audioBuffer.Bytes()
	p.audioBuffer.Reset()
	p.bufferMu.Unlock()

	if len(audio) == 0 {
		p.log.Warn().Msg("No audio recorded")
		return
	}

	// Check minimum speech duration
	minSpeech := time.Duration(p.cfg.Audio.VAD.MinSpeechMs) * time.Millisecond
	if p.hasSpeech && !p.speechStart.IsZero() {
		speechDuration := time.Since(p.speechStart)
		if speechDuration < minSpeech {
			p.log.Debug().
				Dur("duration", speechDuration).
				Dur("minimum", minSpeech).
				Msg("Speech too short, ignoring")
			return
		}
	}

	p.log.Info().Int("bytes", len(audio)).Msg("Processing recorded audio")

	// Transcribe
	result, err := p.sttProvider.Transcribe(ctx, audio)
	if err != nil {
		p.log.Error().Err(err).Msg("Transcription failed")
		if p.onError != nil {
			p.onError(fmt.Errorf("transcription failed: %w", err))
		}
		return
	}

	text := result.Text
	if text == "" {
		p.log.Debug().Msg("Empty transcription")
		return
	}

	p.log.Info().Str("text", text).Msg("Transcribed")

	// Check for deactivation word first - if detected, exit the session
	if llm.IsAnaDeactivated(text) {
		p.log.Info().Str("text", text).Msg("Deactivation word detected - ending session")
		if p.onResponse != nil {
			p.onResponse("Adiós! Estoy aquí si me necesitas.")
		}
		p.setState(StateIdle)
		return
	}

	// In persistent session (StateListening), we don't require "Ana" prefix for each command
	// Only check for Ana activation if we're starting a new session from Idle state
	currentState := p.GetState()
	if currentState == StateIdle && !llm.IsAnaActivated(text) {
		p.log.Debug().Str("text", text).Msg("Ignoring transcription - Ana name not mentioned")
		return
	}

	if p.onTranscript != nil {
		p.onTranscript(text)
	}

	// Process command
	response, err := p.brain.ProcessCommand(ctx, text)
	if err != nil {
		p.log.Error().Err(err).Msg("Command processing failed")
		if p.onError != nil {
			p.onError(fmt.Errorf("command processing failed: %w", err))
		}
		// Keep session active after error
		p.setState(StateListening)
		return
	}

	if response != "" {
		p.log.Info().Str("response", response).Msg("Response")

		if p.onResponse != nil {
			p.onResponse(response)
		}
	}

	// Speak response
	if err := p.brain.ProcessAndSpeak(ctx, text); err != nil {
		p.log.Error().Err(err).Msg("TTS failed")
	}

	// After processing a command, return to Listening state instead of Idle
	// This keeps the session active for continuous interaction
	p.setState(StateListening)
}
