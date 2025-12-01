# GitHub Copilot Instructions

## Project: AnaStreamer

Local voice assistant for streamers written in Go.

## What This Project Does

1. Listens to voice (wake word "Ana" or hotkey)
2. Transcribes with Whisper.cpp or OpenAI
3. Interprets with Ollama or OpenAI
4. Executes actions (Twitch, OBS, Music)
5. Responds with Piper or OpenAI TTS

## Key Patterns

### Interface-based providers
```go
type Provider interface {
    Name() string
    IsAvailable(ctx context.Context) bool
    Close() error
}
```

### Action struct from LLM
```go
type Action struct {
    Action string                 `json:"action"`
    Params map[string]interface{} `json:"params"`
    Reply  string                 `json:"reply"`
}
```

### Error handling
```go
if err != nil {
    return fmt.Errorf("failed to do X: %w", err)
}
```

### Logging
```go
log := logger.Component("module")
log.Info().Str("key", "value").Msg("message")
```

### Constructor pattern
```go
func NewProvider(cfg Config) (*Provider, error)
```

## Directory Structure

- `cmd/ana/` - Entry point
- `internal/config/` - Configuration
- `internal/stt/` - Speech-to-Text
- `internal/llm/` - Language Model
- `internal/tts/` - Text-to-Speech
- `internal/brain/` - Orchestrator
- `internal/executor/` - Action executors
- `internal/pipeline/` - Processing pipeline
- `pkg/logger/` - Logging
- `pkg/utils/` - Utilities

## Available Actions

- twitch: clip, title, category, ban, timeout, unban
- obs: scene, source.show, source.hide, volume, mute, unmute, text
- music: play, pause, resume, next, previous, volume, stop

## Code Preferences

- Go 1.22+
- Context as first parameter
- Wrap errors with context
- Use zerolog for logging
- Snake_case for files
- CamelCase for exports
