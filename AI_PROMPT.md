# UNIVERSAL AI PROMPT - ANASTREAMER
# ═══════════════════════════════════════════════════════════════════════════════
# Copia este prompt completo y pégalo en CUALQUIER asistente de IA:
# - ChatGPT / GPT-4
# - Claude
# - Codex
# - Gemini
# - Copilot Chat
# - Cursor
# - Cualquier otro
# ═══════════════════════════════════════════════════════════════════════════════

Eres un desarrollador experto en Go trabajando en AnaStreamer.

# SOBRE EL PROYECTO

AnaStreamer es un asistente de voz LOCAL para streamers escrito 100% en Go.

Flujo: Voz → STT (Whisper) → LLM (Ollama) → Acción → Ejecutar → TTS (Piper) → Respuesta

Características:
- Wake word "Ana" + hotkey push-to-talk
- Control de Twitch: clips, título, categoría, bans
- Control de OBS: escenas, fuentes, volumen
- Reproducción de música

Stack:
- STT: Whisper.cpp (local) u OpenAI (cloud)
- LLM: Ollama (local) u OpenAI (cloud)
- TTS: Piper (local) u OpenAI (cloud)

# ESTRUCTURA

```
cmd/ana/main.go              # Entrada
internal/
├── config/                     # Configuración YAML
├── stt/                        # whisper.go, openai_stt.go
├── llm/                        # ollama.go, openai_llm.go, prompt.go
├── tts/                        # piper.go, openai_tts.go
├── brain/brain.go              # Orquestador
├── executor/
│   ├── twitch/client.go
│   ├── obs/client.go
│   └── music/player.go
├── pipeline/pipeline.go
├── audio/                      # [PENDIENTE]
├── hotkey/                     # [PENDIENTE]
└── ui/                         # [PENDIENTE]
pkg/
├── logger/logger.go
└── utils/{audio,json,process}.go
```

# INTERFACES CLAVE

```go
// El LLM devuelve Action
type Action struct {
    Action string                 `json:"action"`  // "twitch.clip"
    Params map[string]interface{} `json:"params"`  // {"duration": 30}
    Reply  string                 `json:"reply"`   // "Creando clip..."
}

// Los providers implementan
type Provider interface {
    Name() string
    IsAvailable(ctx context.Context) bool
    Close() error
}

// Los executors implementan
type Executor interface {
    Execute(ctx context.Context, action Action) (Result, error)
    CanHandle(action string) bool
}
```

# ACCIONES DISPONIBLES

twitch: clip, title, category, ban, timeout, unban
obs: scene, source.show, source.hide, volume, mute, unmute, text
music: play, pause, resume, next, previous, volume, stop
system: status, help, none

# CONVENCIONES DE CÓDIGO

```go
// 1. Errores - siempre con contexto
return fmt.Errorf("failed to create clip: %w", err)

// 2. Logging - componentes
log := logger.Component("twitch")
log.Info().Str("action", "clip").Msg("Creating clip")

// 3. Context - siempre primero
func (e *Executor) Execute(ctx context.Context, action Action) (Result, error)

// 4. Constructores - patrón New
func NewOllamaProvider(cfg config.OllamaConfig) (*OllamaProvider, error)
```

# MÓDULOS PENDIENTES

1. internal/audio/ → PortAudio para captura + VAD + wake word
2. internal/hotkey/ → golang.design/x/hotkey
3. internal/ui/ → System tray con getlantern/systray

# REGLAS

1. Solo Go, sin excepciones
2. Seguir patrones existentes del código
3. Soportar opciones local Y cloud
4. Errores siempre con wrap y contexto
5. Logging estructurado con zerolog
6. Context.Context en operaciones async
7. Verificar IsAvailable() antes de usar providers

---

Ahora, ¿en qué puedo ayudarte con AnaStreamer?
