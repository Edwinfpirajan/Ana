# ANASTREAMER - AI AGENTS CONTEXT
# ═══════════════════════════════════════════════════════════════════════════════
# Este archivo sirve como contexto para CUALQUIER agente de IA de código:
# - OpenAI Codex / ChatGPT
# - Claude (Anthropic)
# - GitHub Copilot
# - Cursor AI
# - Codeium
# - Amazon CodeWhisperer
# - Google Gemini
# ═══════════════════════════════════════════════════════════════════════════════

## IDENTITY

You are an expert Go developer working on AnaStreamer.

## PROJECT SUMMARY

**AnaStreamer** = Local voice assistant for streamers (100% Go)

```
Voice → STT (Whisper) → LLM (Ollama) → Action → Execute → TTS (Piper) → Voice
```

**Core Features:**
- Wake word "Ana" + push-to-talk hotkey
- Twitch control: clips, title, category, bans
- OBS control: scenes, sources, volume
- Music playback: play, pause, next, volume

**Tech Stack:**
- Language: Go 1.22+
- STT: Whisper.cpp (local) | OpenAI (cloud)
- LLM: Ollama (local) | OpenAI (cloud)
- TTS: Piper (local) | OpenAI (cloud)
- APIs: Twitch Helix, OBS WebSocket 5.x

---

## DIRECTORY STRUCTURE

```
AnaStreamer/
├── cmd/ana/main.go              # Entry point
├── internal/
│   ├── config/                     # YAML config (config.go, loader.go, defaults.go)
│   ├── stt/                        # Speech-to-Text
│   │   ├── stt.go                  # Interface
│   │   ├── whisper.go              # Local Whisper.cpp
│   │   └── openai_stt.go           # OpenAI Whisper API
│   ├── llm/                        # Language Model
│   │   ├── llm.go                  # Interface + Action struct
│   │   ├── prompt.go               # System prompts
│   │   ├── ollama.go               # Local Ollama
│   │   └── openai_llm.go           # OpenAI GPT
│   ├── tts/                        # Text-to-Speech
│   │   ├── tts.go                  # Interface
│   │   ├── piper.go                # Local Piper
│   │   └── openai_tts.go           # OpenAI TTS
│   ├── brain/brain.go              # Central orchestrator
│   ├── executor/                   # Action executors
│   │   ├── executor.go             # Interface + Registry
│   │   ├── twitch/client.go        # Twitch Helix API
│   │   ├── obs/client.go           # OBS WebSocket
│   │   └── music/player.go         # Music player
│   ├── pipeline/pipeline.go        # Processing pipeline
│   ├── audio/                      # [TODO] Audio capture
│   ├── hotkey/                     # [TODO] Global hotkeys
│   └── ui/                         # [TODO] System tray
├── pkg/
│   ├── logger/logger.go            # Zerolog wrapper
│   └── utils/                      # Utilities
│       ├── audio.go                # PCM/WAV conversion
│       ├── json.go                 # JSON helpers
│       └── process.go              # Process execution
├── config/ana.config.example.yaml
├── go.mod
└── Makefile
```

---

## KEY INTERFACES

### STT Provider
```go
type Provider interface {
    Name() string
    Transcribe(ctx context.Context, audio []byte) (*TranscriptionResult, error)
    TranscribeFile(ctx context.Context, filePath string) (*TranscriptionResult, error)
    SetLanguage(lang string)
    IsAvailable(ctx context.Context) bool
    Close() error
}

type TranscriptionResult struct {
    Text       string
    Language   string
    Confidence float64
    Duration   float64
}
```

### LLM Provider
```go
type Provider interface {
    Name() string
    Complete(ctx context.Context, prompt string) (Action, error)
    CompleteRaw(ctx context.Context, prompt string) (string, error)
    IsAvailable(ctx context.Context) bool
    Close() error
}

type Action struct {
    Action string                 `json:"action"`  // e.g., "twitch.clip"
    Params map[string]interface{} `json:"params"`  // e.g., {"duration": 30}
    Reply  string                 `json:"reply"`   // e.g., "Creating clip..."
}
```

### TTS Provider
```go
type Provider interface {
    Name() string
    Speak(ctx context.Context, text string) error
    Synthesize(ctx context.Context, text string) ([]byte, error)
    SetVoice(voice string) error
    SetSpeed(speed float64)
    Stop()
    IsAvailable(ctx context.Context) bool
    Close() error
}
```

### Action Executor
```go
type Executor interface {
    Name() string
    SupportedActions() []string
    CanHandle(action string) bool
    Execute(ctx context.Context, action llm.Action) (Result, error)
    IsAvailable() bool
    Close() error
}

type Result struct {
    Success bool
    Message string
    Data    map[string]interface{}
    Error   string
}
```

---

## AVAILABLE ACTIONS

| Domain | Actions |
|--------|---------|
| `twitch.*` | `clip`, `title`, `category`, `ban`, `timeout`, `unban` |
| `obs.*` | `scene`, `source.show`, `source.hide`, `volume`, `mute`, `unmute`, `text` |
| `music.*` | `play`, `pause`, `resume`, `next`, `previous`, `volume`, `stop` |
| `system.*` | `status`, `help` |
| `none` | No action (conversation only) |

**Action JSON Format:**
```json
{"action": "twitch.clip", "params": {"duration": 30}, "reply": "Creating 30s clip"}
```

---

## CODE CONVENTIONS

```go
// 1. Error wrapping - ALWAYS wrap with context
if err != nil {
    return fmt.Errorf("failed to create clip: %w", err)
}

// 2. Logging - Use component loggers
log := logger.Component("twitch")
log.Info().Str("action", "clip").Msg("Creating clip")

// 3. Context - Always first parameter
func (e *Executor) Execute(ctx context.Context, action Action) (Result, error)

// 4. Constructors - New...() pattern
func NewOllamaProvider(cfg config.OllamaConfig) (*OllamaProvider, error)

// 5. Interfaces - Check availability before use
if !provider.IsAvailable(ctx) {
    return fmt.Errorf("provider not available")
}
```

**File Naming:** `snake_case.go` (e.g., `openai_stt.go`)

---

## CONFIGURATION (ana.config.yaml)

```yaml
general:
  language: "es"
  log_level: "info"

stt:
  provider: "whisper"  # whisper | openai
  whisper:
    binary_path: "./bin/whisper"
    model_path: "./assets/models/whisper/ggml-base.bin"
    language: "es"

llm:
  provider: "ollama"  # ollama | openai
  ollama:
    url: "http://localhost:11434"
    model: "llama3.2:3b"
    timeout_seconds: 30

tts:
  provider: "piper"  # piper | openai
  piper:
    binary_path: "./bin/piper"
    model_path: "./assets/voices/piper/es_ES-davefx-medium.onnx"

twitch:
  enabled: true
  client_id: "xxx"
  broadcaster_id: "xxx"

obs:
  enabled: true
  url: "ws://localhost:4455"

music:
  enabled: true
  folders: ["./music"]
```

---

## EXTERNAL INTEGRATIONS

| Service | Protocol | Notes |
|---------|----------|-------|
| Twitch | REST (Helix API) | OAuth Bearer + Client-ID headers |
| OBS | WebSocket 5.x | SHA256 challenge auth |
| Ollama | HTTP | `POST /api/generate` with `format: "json"` |
| Whisper.cpp | CLI | Binary execution with args |
| Piper | CLI | Text via stdin, WAV via stdout |

---

## PENDING MODULES (TODO)

### 1. internal/audio/
```go
// Need to implement:
type AudioCapturer interface {
    Start() error
    Stop() error
    AudioStream() <-chan []byte
}

type VADDetector interface {
    Process(samples []byte) VADResult
    SetSensitivity(level float64)
}

type WakeWordDetector interface {
    Process(samples []byte) bool
    OnDetected() <-chan WakeEvent
}
```
**Library:** `github.com/gordonklaus/portaudio`

### 2. internal/hotkey/
```go
type HotkeyListener interface {
    Register(key string, callback func()) error
    Start() error
    Stop()
}
```
**Library:** `golang.design/x/hotkey`

### 3. internal/ui/
```go
// System tray with status and controls
```
**Library:** `github.com/getlantern/systray`

---

## BUILD & RUN

```bash
# Build
go build -o ana ./cmd/ana

# Run interactive
./ana

# Test command
./ana -test -command "create a clip"

# Check status
Ana> status
```

---

## RULES FOR AI AGENTS

1. **Language:** Go only, no exceptions
2. **Patterns:** Follow existing code patterns
3. **Errors:** Always wrap with `fmt.Errorf("context: %w", err)`
4. **Logging:** Use `logger.Component("name")`
5. **Context:** Pass `context.Context` as first param
6. **Config:** All settings via `*config.Config`
7. **Providers:** Support both local and cloud options
8. **Testing:** Add tests for new functionality
