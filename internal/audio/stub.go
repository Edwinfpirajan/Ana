//go:build !portaudio

package audio

import (
	"context"
	"fmt"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/internal/pipeline"
	"github.com/anastreamer/ana/internal/stt"
)

type capture struct{}

// Start returns a stub error when PortAudio is disabled.
func Start(ctx context.Context, cfg config.AudioConfig, pip *pipeline.Pipeline, sttProvider stt.Provider, onWake func()) (*capture, error) {
	return nil, fmt.Errorf("PortAudio build tag is required for audio capture")
}

func (c *capture) Stop() {}
