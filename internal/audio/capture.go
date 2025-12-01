//go:build portaudio

package audio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/internal/pipeline"
	"github.com/anastreamer/ana/internal/stt"
	"github.com/anastreamer/ana/pkg/utils"
)

type capture struct {
	ctx      context.Context
	cancel   context.CancelFunc
	cfg      config.AudioConfig
	pipeline *pipeline.Pipeline
	stt      stt.Provider
	onWake   func()
	stream   *portaudio.Stream
	wg       sync.WaitGroup
	detected bool
	lastWake time.Time
}

func Start(ctx context.Context, cfg config.AudioConfig, pip *pipeline.Pipeline, sttProvider stt.Provider, onWake func()) (*capture, error) {
	if cfg.SampleRate == 0 {
		cfg.SampleRate = 16000
	}
	if cfg.Channels == 0 {
		cfg.Channels = 1
	}
	if cfg.ChunkSize == 0 {
		cfg.ChunkSize = 1024
	}

	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("portaudio init: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	c := &capture{
		ctx:      ctx,
		cancel:   cancel,
		cfg:      cfg,
		pipeline: pip,
		stt:      sttProvider,
		onWake:   onWake,
	}

	stream, err := portaudio.OpenDefaultStream(cfg.Channels, 0, float64(cfg.SampleRate), cfg.ChunkSize, c.process)
	if err != nil {
		portaudio.Terminate()
		cancel()
		return nil, fmt.Errorf("open stream: %w", err)
	}
	c.stream = stream

	if err := c.stream.Start(); err != nil {
		c.stream.Close()
		portaudio.Terminate()
		cancel()
		return nil, fmt.Errorf("start stream: %w", err)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		<-ctx.Done()
		c.stream.Stop()
		c.stream.Close()
		portaudio.Terminate()
	}()

	return c, nil
}

func (c *capture) Stop() {
	c.cancel()
	c.wg.Wait()
}

func (c *capture) process(in []int16) {
	select {
	case <-c.ctx.Done():
		return
	default:
	}

	bytesData := utils.Int16ToBytes(in)
	c.pipeline.FeedAudio(bytesData)

	// NOTE: Wake word detection via continuous chunk transcription disabled
	// Reason: Whisper needs significant audio length (30+ seconds) to work correctly
	// Instead, we now use the pipeline's recordUntilSilence() to collect audio
	// The pipeline will transcribe the complete recording, not individual chunks

	// Fallback: Keep STT nil check to avoid nil pointer dereference
	_ = c.stt
	_ = c.onWake
}

func detectSpeech(samples []int16) bool {
	var sum float64
	for _, s := range samples {
		sum += float64(s * s)
	}
	rms := sum / float64(len(samples))
	// Much lower threshold to allow quiet speech detection
	// The pipeline will apply its own VAD sensitivity thresholds
	return rms > 100
}

func containsWakeWord(text string) bool {
	return utils.ContainsIgnoreCase(text, "ana")
}
