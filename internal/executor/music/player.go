// Package music provides music playback for AnaStreamer
package music

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/internal/executor"
	"github.com/anastreamer/ana/internal/llm"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/rs/zerolog"
)

// Executor implements the music player executor
type Executor struct {
	folders          []string
	supportedFormats []string
	defaultVolume    float64
	shuffle          bool
	log              zerolog.Logger
	enabled          bool

	mu         sync.Mutex
	playlist   []string
	currentIdx int
	isPlaying  bool
	isPaused   bool
	volume     float64
	currentCmd *exec.Cmd
	stopChan   chan struct{}
}

// NewExecutor creates a new music executor
func NewExecutor(cfg config.MusicConfig) *Executor {
	return &Executor{
		folders:          cfg.Folders,
		supportedFormats: cfg.SupportedFormats,
		defaultVolume:    cfg.DefaultVolume,
		shuffle:          cfg.Shuffle,
		log:              logger.Component("music"),
		enabled:          cfg.Enabled,
		volume:           cfg.DefaultVolume,
		stopChan:         make(chan struct{}),
	}
}

// Name returns the executor name
func (e *Executor) Name() string {
	return "music"
}

// SupportedActions returns the list of supported actions
func (e *Executor) SupportedActions() []string {
	return []string{
		"music.play",
		"music.pause",
		"music.resume",
		"music.next",
		"music.previous",
		"music.volume",
		"music.stop",
	}
}

// CanHandle returns true if this executor can handle the action
func (e *Executor) CanHandle(action string) bool {
	return strings.HasPrefix(action, "music.")
}

// Execute executes a music action
func (e *Executor) Execute(ctx context.Context, action llm.Action) (executor.Result, error) {
	if !e.enabled {
		return executor.NewErrorResult(fmt.Errorf("music is not enabled")), nil
	}

	switch action.Action {
	case "music.play":
		return e.play(ctx, action)
	case "music.pause":
		return e.pause(ctx)
	case "music.resume":
		return e.resume(ctx)
	case "music.next":
		return e.next(ctx)
	case "music.previous":
		return e.previous(ctx)
	case "music.volume":
		return e.setVolume(ctx, action)
	case "music.stop":
		return e.stop(ctx)
	default:
		return executor.NewErrorResult(fmt.Errorf("unknown music action: %s", action.Action)), nil
	}
}

// IsAvailable checks if music is available
func (e *Executor) IsAvailable() bool {
	return e.enabled
}

// Close releases resources
func (e *Executor) Close() error {
	e.stopPlayback()
	return nil
}

// play starts playing music
func (e *Executor) play(ctx context.Context, action llm.Action) (executor.Result, error) {
	query := action.GetStringParam("query")

	// Load playlist
	if err := e.loadPlaylist(query); err != nil {
		return executor.NewErrorResult(err), err
	}

	if len(e.playlist) == 0 {
		return executor.NewErrorResult(fmt.Errorf("no music found")), nil
	}

	// Shuffle if enabled
	if e.shuffle {
		e.shufflePlaylist()
	}

	// Start playing
	e.currentIdx = 0
	go e.playLoop()

	currentTrack := filepath.Base(e.playlist[e.currentIdx])
	e.log.Info().Str("track", currentTrack).Int("total", len(e.playlist)).Msg("Playing music")

	return executor.NewResultWithData("Playing music", map[string]interface{}{
		"track": currentTrack,
		"total": len(e.playlist),
	}), nil
}

// loadPlaylist loads music files from configured folders
func (e *Executor) loadPlaylist(query string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.playlist = nil
	query = strings.ToLower(query)

	for _, folder := range e.folders {
		err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if info.IsDir() {
				return nil
			}

			// Check if supported format
			ext := strings.ToLower(filepath.Ext(path))
			supported := false
			for _, fmt := range e.supportedFormats {
				if ext == fmt {
					supported = true
					break
				}
			}
			if !supported {
				return nil
			}

			// Filter by query if provided
			if query != "" {
				name := strings.ToLower(filepath.Base(path))
				if !strings.Contains(name, query) {
					return nil
				}
			}

			e.playlist = append(e.playlist, path)
			return nil
		})
		if err != nil {
			e.log.Warn().Err(err).Str("folder", folder).Msg("Error scanning folder")
		}
	}

	return nil
}

// shufflePlaylist shuffles the playlist
func (e *Executor) shufflePlaylist() {
	e.mu.Lock()
	defer e.mu.Unlock()

	rand.Shuffle(len(e.playlist), func(i, j int) {
		e.playlist[i], e.playlist[j] = e.playlist[j], e.playlist[i]
	})
}

// playLoop is the main playback loop
func (e *Executor) playLoop() {
	for {
		e.mu.Lock()
		if e.currentIdx >= len(e.playlist) {
			e.isPlaying = false
			e.mu.Unlock()
			return
		}
		track := e.playlist[e.currentIdx]
		e.isPlaying = true
		e.isPaused = false
		e.mu.Unlock()

		// Play the track
		err := e.playTrack(track)
		if err != nil {
			e.log.Error().Err(err).Str("track", track).Msg("Error playing track")
		}

		// Check if we should continue
		select {
		case <-e.stopChan:
			e.mu.Lock()
			e.isPlaying = false
			e.mu.Unlock()
			return
		default:
		}

		// Move to next track
		e.mu.Lock()
		e.currentIdx++
		e.mu.Unlock()
	}
}

// playTrack plays a single track
func (e *Executor) playTrack(path string) error {
	e.log.Debug().Str("track", path).Msg("Playing track")

	// Find available player
	var cmd *exec.Cmd

	players := []struct {
		name string
		args func(string, float64) []string
	}{
		{"mpv", func(path string, vol float64) []string {
			return []string{"--no-video", "--really-quiet", fmt.Sprintf("--volume=%.0f", vol*100), path}
		}},
		{"ffplay", func(path string, vol float64) []string {
			return []string{"-nodisp", "-autoexit", "-loglevel", "quiet", "-volume", fmt.Sprintf("%.0f", vol*100), path}
		}},
		{"afplay", func(path string, vol float64) []string {
			return []string{"-v", fmt.Sprintf("%.2f", vol), path}
		}},
	}

	for _, player := range players {
		if _, err := exec.LookPath(player.name); err == nil {
			args := player.args(path, e.volume)
			cmd = exec.Command(player.name, args...)
			break
		}
	}

	if cmd == nil {
		return fmt.Errorf("no audio player found (install mpv or ffplay)")
	}

	e.mu.Lock()
	e.currentCmd = cmd
	e.mu.Unlock()

	err := cmd.Run()

	e.mu.Lock()
	e.currentCmd = nil
	e.mu.Unlock()

	return err
}

// pause pauses playback
func (e *Executor) pause(ctx context.Context) (executor.Result, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isPlaying {
		return executor.NewErrorResult(fmt.Errorf("nothing is playing")), nil
	}

	if e.currentCmd != nil && e.currentCmd.Process != nil {
		// Send SIGSTOP on Unix systems
		// For cross-platform, we'll just kill and remember position
		e.currentCmd.Process.Kill()
	}
	e.isPaused = true

	e.log.Info().Msg("Music paused")
	return executor.NewResult("Music paused"), nil
}

// resume resumes playback
func (e *Executor) resume(ctx context.Context) (executor.Result, error) {
	e.mu.Lock()

	if !e.isPaused {
		e.mu.Unlock()
		return executor.NewErrorResult(fmt.Errorf("music is not paused")), nil
	}

	e.isPaused = false
	e.mu.Unlock()

	// Restart playback loop
	go e.playLoop()

	e.log.Info().Msg("Music resumed")
	return executor.NewResult("Music resumed"), nil
}

// next skips to the next track
func (e *Executor) next(ctx context.Context) (executor.Result, error) {
	e.mu.Lock()

	if len(e.playlist) == 0 {
		e.mu.Unlock()
		return executor.NewErrorResult(fmt.Errorf("no playlist")), nil
	}

	// Kill current track
	if e.currentCmd != nil && e.currentCmd.Process != nil {
		e.currentCmd.Process.Kill()
	}

	// Index will be incremented by playLoop
	nextIdx := e.currentIdx + 1
	if nextIdx >= len(e.playlist) {
		nextIdx = 0 // Loop back
	}

	track := ""
	if nextIdx < len(e.playlist) {
		track = filepath.Base(e.playlist[nextIdx])
	}

	e.mu.Unlock()

	e.log.Info().Str("track", track).Msg("Next track")
	return executor.NewResultWithData("Next track", map[string]interface{}{
		"track": track,
	}), nil
}

// previous goes to the previous track
func (e *Executor) previous(ctx context.Context) (executor.Result, error) {
	e.mu.Lock()

	if len(e.playlist) == 0 {
		e.mu.Unlock()
		return executor.NewErrorResult(fmt.Errorf("no playlist")), nil
	}

	// Kill current track
	if e.currentCmd != nil && e.currentCmd.Process != nil {
		e.currentCmd.Process.Kill()
	}

	// Go back one (will be incremented by playLoop, so go back 2)
	e.currentIdx -= 2
	if e.currentIdx < 0 {
		e.currentIdx = len(e.playlist) - 2
	}
	if e.currentIdx < 0 {
		e.currentIdx = 0
	}

	track := filepath.Base(e.playlist[e.currentIdx+1])
	e.mu.Unlock()

	e.log.Info().Str("track", track).Msg("Previous track")
	return executor.NewResultWithData("Previous track", map[string]interface{}{
		"track": track,
	}), nil
}

// setVolume changes the volume
func (e *Executor) setVolume(ctx context.Context, action llm.Action) (executor.Result, error) {
	volume := action.GetFloatParam("volume")
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	e.mu.Lock()
	e.volume = volume
	e.mu.Unlock()

	e.log.Info().Float64("volume", volume).Msg("Volume set")
	return executor.NewResult(fmt.Sprintf("Volume set to %.0f%%", volume*100)), nil
}

// stop stops playback
func (e *Executor) stop(ctx context.Context) (executor.Result, error) {
	e.stopPlayback()

	e.log.Info().Msg("Music stopped")
	return executor.NewResult("Music stopped"), nil
}

// stopPlayback stops the current playback
func (e *Executor) stopPlayback() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.currentCmd != nil && e.currentCmd.Process != nil {
		e.currentCmd.Process.Kill()
	}

	// Signal stop
	select {
	case e.stopChan <- struct{}{}:
	default:
	}

	e.isPlaying = false
	e.isPaused = false
}

// GetCurrentTrack returns the currently playing track
func (e *Executor) GetCurrentTrack() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isPlaying || e.currentIdx >= len(e.playlist) {
		return ""
	}
	return filepath.Base(e.playlist[e.currentIdx])
}

// IsPlaying returns whether music is currently playing
func (e *Executor) IsPlaying() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.isPlaying && !e.isPaused
}
