# CODEX INSTRUCTIONS - ANASTREAMER

> **For OpenAI Codex, ChatGPT, and GPT-4 models**

## System Prompt

You are an expert Go developer. You are working on AnaStreamer, a local voice assistant for streamers written in Go.

## Project Facts

- **Name:** AnaStreamer
- **Language:** Go 1.22+ (only Go, no other languages)
- **Purpose:** Voice assistant for streamers (Twitch, OBS, Music)
- **Architecture:** Modular with interfaces for swappable providers
- **Mode:** 100% local-first (can optionally use cloud APIs)

## Tech Stack

| Component | Local Option | Cloud Option |
|-----------|--------------|--------------|
| STT | Whisper.cpp | OpenAI Whisper |
| LLM | Ollama | OpenAI GPT |
| TTS | Piper | OpenAI TTS |

## Core Flow

```
User Voice → AudioCapture → STT → LLM → Action JSON → Executor → TTS → Response
```

## Directory Map

```
cmd/ana/main.go           = Entry point
internal/config/             = Config structs + YAML loader
internal/stt/                = Speech-to-Text (whisper.go, openai_stt.go)
internal/llm/                = LLM + Action struct (ollama.go, openai_llm.go)
internal/tts/                = Text-to-Speech (piper.go, openai_tts.go)
internal/brain/              = Orchestrator (brain.go)
internal/executor/twitch/    = Twitch API client
internal/executor/obs/       = OBS WebSocket client
internal/executor/music/     = Music player
internal/pipeline/           = Audio processing pipeline
pkg/logger/                  = Zerolog wrapper
pkg/utils/                   = Helpers (audio, json, process)
```

## Key Types

```go
// LLM returns this
type Action struct {
    Action string                 `json:"action"`  // "twitch.clip"
    Params map[string]interface{} `json:"params"`  // {"duration": 30}
    Reply  string                 `json:"reply"`   // "Creating clip..."
}

// Executor returns this
type Result struct {
    Success bool
    Message string
    Data    map[string]interface{}
    Error   string
}
```

## Interface Pattern

All providers follow:
```go
type Provider interface {
    Name() string
    IsAvailable(ctx context.Context) bool
    Close() error
    // + domain-specific methods
}
```

## Actions List

```
twitch.clip, twitch.title, twitch.category, twitch.ban, twitch.timeout, twitch.unban
obs.scene, obs.source.show, obs.source.hide, obs.volume, obs.mute, obs.unmute, obs.text
music.play, music.pause, music.resume, music.next, music.previous, music.volume, music.stop
system.status, system.help, none
```

## Code Style

```go
// Errors: wrap with context
return fmt.Errorf("failed to X: %w", err)

// Logging: component-based
log := logger.Component("name")
log.Info().Str("k", "v").Msg("message")

// Constructors
func NewSomething(cfg Config) (*Something, error)

// Context first
func (x *X) Method(ctx context.Context, ...) error
```

## TODO Modules

1. `internal/audio/` - PortAudio capture + VAD + wake word
2. `internal/hotkey/` - Global hotkeys (golang.design/x/hotkey)  
3. `internal/ui/` - System tray (getlantern/systray)

## When Generating Code

1. Follow existing patterns
2. Use interfaces for swappable components
3. Support both local and cloud providers
4. Handle errors with wrapped context
5. Use structured logging
6. Add context.Context to async operations
7. Check IsAvailable() before using providers

## Example Task

"Implement the audio capture module with PortAudio"

Response should:
- Create `internal/audio/capture.go` with interface
- Create `internal/audio/portaudio.go` with implementation
- Create `internal/audio/vad.go` for voice detection
- Follow existing code patterns
- Include proper error handling
- Add logging
