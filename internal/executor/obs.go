package executor

import (
	"context"
	"fmt"
	"strings"

	"github.com/anastreamer/ana/internal/config"
	"github.com/anastreamer/ana/internal/llm"
	"github.com/anastreamer/ana/pkg/logger"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/requests/stream"
	"github.com/rs/zerolog"
)

// OBSExecutor handles OBS Studio integration via WebSocket
type OBSExecutor struct {
	cfg    config.OBSConfig
	client *goobs.Client
	log    zerolog.Logger
}

// NewOBSExecutor creates a new OBS executor
func NewOBSExecutor(cfg config.OBSConfig) (*OBSExecutor, error) {
	exec := &OBSExecutor{
		cfg: cfg,
		log: logger.Component("obs-executor"),
	}

	if cfg.Enabled {
		if err := exec.connect(); err != nil {
			exec.log.Warn().Err(err).Msg("Failed to connect to OBS, executor will be unavailable")
			// Don't return error - executor can still be registered but will be unavailable
		}
	}

	return exec, nil
}

// connect establishes connection to OBS WebSocket
func (e *OBSExecutor) connect() error {
	e.log.Info().Str("url", e.cfg.URL).Msg("Connecting to OBS WebSocket")

	// Parse host and port from URL
	host := strings.TrimPrefix(e.cfg.URL, "ws://")
	host = strings.TrimPrefix(host, "wss://")

	e.log.Debug().Str("host", host).Msg("Parsed host for OBS connection")

	var client *goobs.Client
	var err error

	// Connect with or without password
	if e.cfg.Password != "" {
		client, err = goobs.New(host, goobs.WithPassword(e.cfg.Password))
	} else {
		client, err = goobs.New(host)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to OBS: %w", err)
	}

	e.client = client
	e.log.Info().Msg("Successfully connected to OBS WebSocket")
	return nil
}

// Name returns the executor name
func (e *OBSExecutor) Name() string {
	return "obs"
}

// SupportedActions returns all supported OBS actions
func (e *OBSExecutor) SupportedActions() []string {
	return []string{
		"obs.start_recording",
		"obs.stop_recording",
		"obs.start_streaming",
		"obs.stop_streaming",
		"obs.scene",
	}
}

// CanHandle checks if this executor can handle the given action
func (e *OBSExecutor) CanHandle(action string) bool {
	return strings.HasPrefix(action, "obs.")
}

// Execute executes an OBS action
func (e *OBSExecutor) Execute(ctx context.Context, action llm.Action) (Result, error) {
	if !e.IsAvailable() {
		return NewErrorResult(fmt.Errorf("OBS not connected")), fmt.Errorf("OBS not connected")
	}

	e.log.Debug().Str("action", action.Action).Msg("Executing OBS action")

	switch action.Action {
	case "obs.start_recording":
		return e.startRecording(ctx)
	case "obs.stop_recording":
		return e.stopRecording(ctx)
	case "obs.start_streaming":
		return e.startStreaming(ctx)
	case "obs.stop_streaming":
		return e.stopStreaming(ctx)
	case "obs.scene":
		return e.switchScene(ctx, action)
	default:
		return NewErrorResult(fmt.Errorf("unknown action: %s", action.Action)), nil
	}
}

// startRecording starts OBS recording
func (e *OBSExecutor) startRecording(ctx context.Context) (Result, error) {
	e.log.Info().Msg("Starting OBS recording")

	// Check if already recording
	status, err := e.client.Record.GetRecordStatus()
	if err != nil {
		e.log.Error().Err(err).Msg("Failed to get record status")
		return NewErrorResult(err), err
	}

	if status.OutputActive {
		e.log.Warn().Msg("Recording is already active")
		return NewResult("La grabación ya está activa"), nil
	}

	_, err = e.client.Record.StartRecord()
	if err != nil {
		e.log.Error().Err(err).Msg("Failed to start recording")
		return NewErrorResult(err), err
	}

	e.log.Info().Msg("Recording started successfully")
	return NewResult("Grabación iniciada"), nil
}

// stopRecording stops OBS recording
func (e *OBSExecutor) stopRecording(ctx context.Context) (Result, error) {
	e.log.Info().Msg("Stopping OBS recording")

	// Check if recording is active
	status, err := e.client.Record.GetRecordStatus()
	if err != nil {
		e.log.Error().Err(err).Msg("Failed to get record status")
		return NewErrorResult(err), err
	}

	if !status.OutputActive {
		e.log.Warn().Msg("Recording is not active")
		return NewResult("La grabación ya está detenida"), nil
	}

	resp, err := e.client.Record.StopRecord()
	if err != nil {
		e.log.Error().Err(err).Msg("Failed to stop recording")
		return NewErrorResult(err), err
	}

	e.log.Info().Str("output_path", resp.OutputPath).Msg("Recording stopped successfully")
	return NewResult(fmt.Sprintf("Grabación detenida. Archivo: %s", resp.OutputPath)), nil
}

// startStreaming starts OBS streaming
func (e *OBSExecutor) startStreaming(ctx context.Context) (Result, error) {
	e.log.Info().Msg("Starting OBS streaming")

	// Check if already streaming
	status, err := e.client.Stream.GetStreamStatus()
	if err != nil {
		e.log.Error().Err(err).Msg("Failed to get stream status")
		return NewErrorResult(err), err
	}

	if status.OutputActive {
		e.log.Warn().Msg("Streaming is already active")
		return NewResult("La transmisión ya está activa"), nil
	}

	_, err = e.client.Stream.StartStream(&stream.StartStreamParams{})
	if err != nil {
		e.log.Error().Err(err).Msg("Failed to start streaming")
		return NewErrorResult(err), err
	}

	e.log.Info().Msg("Streaming started successfully")
	return NewResult("Transmisión iniciada"), nil
}

// stopStreaming stops OBS streaming
func (e *OBSExecutor) stopStreaming(ctx context.Context) (Result, error) {
	e.log.Info().Msg("Stopping OBS streaming")

	// Check if streaming is active
	status, err := e.client.Stream.GetStreamStatus()
	if err != nil {
		e.log.Error().Err(err).Msg("Failed to get stream status")
		return NewErrorResult(err), err
	}

	if !status.OutputActive {
		e.log.Warn().Msg("Streaming is not active")
		return NewResult("La transmisión ya está detenida"), nil
	}

	_, err = e.client.Stream.StopStream()
	if err != nil {
		e.log.Error().Err(err).Msg("Failed to stop streaming")
		return NewErrorResult(err), err
	}

	e.log.Info().Msg("Streaming stopped successfully")
	return NewResult("Transmisión detenida"), nil
}

// switchScene switches to a different OBS scene
func (e *OBSExecutor) switchScene(ctx context.Context, action llm.Action) (Result, error) {
	requestedScene, ok := action.Params["scene"].(string)
	if !ok || requestedScene == "" {
		e.log.Error().Msg("Scene name not provided")
		return NewErrorResult(fmt.Errorf("scene name is required")), fmt.Errorf("scene name is required")
	}

	e.log.Info().Str("requested_scene", requestedScene).Msg("Switching to scene")

	// Get list of available scenes to find exact match
	sceneList, err := e.client.Scenes.GetSceneList()
	if err != nil {
		e.log.Error().Err(err).Msg("Failed to get scene list")
		return NewErrorResult(err), err
	}

	// Log available scenes for debugging
	var sceneNames []string
	for _, scene := range sceneList.Scenes {
		sceneNames = append(sceneNames, scene.SceneName)
	}
	e.log.Debug().Strs("available_scenes", sceneNames).Msg("Available scenes in OBS")

	// Try to find exact match first, then case-insensitive match
	var matchedScene string
	requestedLower := strings.ToLower(requestedScene)

	for _, scene := range sceneList.Scenes {
		sceneName := scene.SceneName
		if sceneName == requestedScene {
			matchedScene = sceneName
			break
		}
		if strings.ToLower(sceneName) == requestedLower {
			matchedScene = sceneName
		}
	}

	if matchedScene == "" {
		e.log.Error().
			Str("requested", requestedScene).
			Strs("available", sceneNames).
			Msg("Scene not found")
		return NewErrorResult(fmt.Errorf("escena '%s' no encontrada. Escenas disponibles: %v", requestedScene, sceneNames)),
			fmt.Errorf("scene not found: %s", requestedScene)
	}

	e.log.Info().Str("matched_scene", matchedScene).Msg("Found matching scene")

	params := scenes.NewSetCurrentProgramSceneParams().WithSceneName(matchedScene)
	_, err = e.client.Scenes.SetCurrentProgramScene(params)
	if err != nil {
		e.log.Error().Err(err).Str("scene", matchedScene).Msg("Failed to switch scene")
		return NewErrorResult(err), err
	}

	e.log.Info().Str("scene", matchedScene).Msg("Scene switched successfully")
	return NewResult(fmt.Sprintf("Cambiando a escena %s", matchedScene)), nil
}

// IsAvailable checks if OBS is connected and ready
func (e *OBSExecutor) IsAvailable() bool {
	return e.cfg.Enabled && e.client != nil
}

// Close disconnects from OBS
func (e *OBSExecutor) Close() error {
	if e.client != nil {
		e.log.Info().Msg("Disconnecting from OBS WebSocket")
		e.client.Disconnect()
		e.client = nil
	}
	return nil
}
