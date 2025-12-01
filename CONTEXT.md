# ANA STREAMER - PROJECT CONTEXT
# ═══════════════════════════════════════════════════════════════════════════════
# Este archivo contiene todo el contexto necesario para entender, modificar
# y extender el proyecto AnaStreamer. Úsalo como referencia o como prompt
# para asistentes de IA.
# ═══════════════════════════════════════════════════════════════════════════════

## 🎯 RESUMEN EJECUTIVO

AnaStreamer es un asistente de voz 100% local para streamers, escrito en Go.
Escucha comandos de voz, los interpreta con un LLM, y ejecuta acciones en
Twitch, OBS y reproduce música. Todo corre localmente sin servidores externos.

**Stack tecnológico:**
- Lenguaje: Go 1.22+
- STT: Whisper.cpp (local) o OpenAI Whisper (cloud)
- LLM: Ollama (local) o OpenAI GPT (cloud)
- TTS: Piper (local) o OpenAI TTS (cloud)
- Integraciones: Twitch Helix API, OBS WebSocket 5.x

---

## 📁 ESTRUCTURA DEL PROYECTO

```
AnaStreamer/
├── cmd/ana/main.go          # Punto de entrada principal
├── internal/
│   ├── config/                  # Configuración YAML
│   │   ├── config.go           # Structs de configuración
│   │   ├── loader.go           # Carga y validación
│   │   └── defaults.go         # Valores por defecto
│   ├── audio/                   # [PENDIENTE] Captura de audio
│   ├── hotkey/                  # [PENDIENTE] Hotkeys del sistema
│   ├── stt/                     # Speech-to-Text
│   │   ├── stt.go              # Interface Provider
│   │   ├── whisper.go          # Implementación Whisper local
│   │   └── openai_stt.go       # Implementación OpenAI
│   ├── llm/                     # Language Model
│   │   ├── llm.go              # Interface Provider + Action struct
│   │   ├── prompt.go           # System prompts
│   │   ├── ollama.go           # Implementación Ollama
│   │   └── openai_llm.go       # Implementación OpenAI
│   ├── tts/                     # Text-to-Speech
│   │   ├── tts.go              # Interface Provider
│   │   ├── piper.go            # Implementación Piper local
│   │   └── openai_tts.go       # Implementación OpenAI
│   ├── brain/                   # Orquestador central
│   │   └── brain.go            # Procesa comandos y despacha acciones
│   ├── executor/                # Ejecutores de acciones
│   │   ├── executor.go         # Interface + Registry
│   │   ├── twitch/client.go    # Acciones de Twitch
│   │   ├── obs/client.go       # Acciones de OBS
│   │   └── music/player.go     # Reproductor de música
│   ├── pipeline/                # Pipeline de procesamiento
│   │   └── pipeline.go         # Orquesta audio→STT→LLM→acciones→TTS
│   └── ui/                      # [PENDIENTE] Interfaz de usuario
├── pkg/
│   ├── logger/logger.go        # Logger estructurado (zerolog)
│   └── utils/                   # Utilidades
│       ├── audio.go            # Conversión PCM↔WAV
│       ├── json.go             # Parsing JSON
│       └── process.go          # Ejecución de procesos externos
├── config/
│   └── ana.config.example.yaml  # Configuración de ejemplo
├── assets/
│   ├── models/whisper/         # Modelos Whisper (.bin)
│   ├── voices/piper/           # Voces Piper (.onnx)
│   └── sounds/                 # Sonidos del sistema
└── bin/                        # Binarios externos (whisper, piper)
```

---

## 🔌 INTERFACES PRINCIPALES

### 1. STT Provider (internal/stt/stt.go)
```go
type Provider interface {
    Name() string
    Transcribe(ctx context.Context, audio []byte) (*TranscriptionResult, error)
    TranscribeFile(ctx context.Context, filePath string) (*TranscriptionResult, error)
    SetLanguage(lang string)
    IsAvailable(ctx context.Context) bool
    Close() error
}
```

### 2. LLM Provider (internal/llm/llm.go)
```go
type Provider interface {
    Name() string
    Complete(ctx context.Context, prompt string) (Action, error)
    CompleteRaw(ctx context.Context, prompt string) (string, error)
    IsAvailable(ctx context.Context) bool
    Close() error
}

type Action struct {
    Action string                 `json:"action"`  // ej: "twitch.clip"
    Params map[string]interface{} `json:"params"`  // ej: {"duration": 30}
    Reply  string                 `json:"reply"`   // ej: "Creando clip..."
}
```

### 3. TTS Provider (internal/tts/tts.go)
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

### 4. Action Executor (internal/executor/executor.go)
```go
type Executor interface {
    Name() string
    SupportedActions() []string
    CanHandle(action string) bool
    Execute(ctx context.Context, action llm.Action) (Result, error)
    IsAvailable() bool
    Close() error
}
```

---

## 🎬 ACCIONES DISPONIBLES

| Dominio | Action | Params | Descripción |
|---------|--------|--------|-------------|
| **Twitch** | `twitch.clip` | `{duration?: int}` | Crear clip |
| | `twitch.title` | `{title: string}` | Cambiar título |
| | `twitch.category` | `{category: string}` | Cambiar categoría |
| | `twitch.ban` | `{user: string, reason?: string}` | Banear usuario |
| | `twitch.timeout` | `{user: string, duration: int}` | Timeout usuario |
| | `twitch.unban` | `{user: string}` | Desbanear |
| **OBS** | `obs.scene` | `{scene: string}` | Cambiar escena |
| | `obs.source.show` | `{source: string}` | Mostrar fuente |
| | `obs.source.hide` | `{source: string}` | Ocultar fuente |
| | `obs.volume` | `{source: string, volume: float}` | Volumen (0-1) |
| | `obs.mute` | `{source: string}` | Mutear |
| | `obs.unmute` | `{source: string}` | Desmutear |
| | `obs.text` | `{source: string, text: string}` | Cambiar texto |
| **Music** | `music.play` | `{query?: string}` | Reproducir |
| | `music.pause` | `{}` | Pausar |
| | `music.resume` | `{}` | Reanudar |
| | `music.next` | `{}` | Siguiente |
| | `music.previous` | `{}` | Anterior |
| | `music.volume` | `{volume: float}` | Volumen (0-1) |
| | `music.stop` | `{}` | Detener |
| **System** | `system.status` | `{}` | Estado |
| | `system.help` | `{}` | Ayuda |
| | `none` | `{}` | Sin acción |

---

## 🔄 FLUJO DE PROCESAMIENTO

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         FLUJO PRINCIPAL                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  [Usuario habla] ──► [Audio Capture] ──► [VAD detecta fin de habla]    │
│                                                │                        │
│                                                ▼                        │
│                                    ┌───────────────────┐                │
│                                    │   STT Provider    │                │
│                                    │ (Whisper/OpenAI)  │                │
│                                    └─────────┬─────────┘                │
│                                              │                          │
│                                    texto: "crea un clip"               │
│                                              │                          │
│                                              ▼                          │
│                                    ┌───────────────────┐                │
│                                    │   LLM Provider    │                │
│                                    │ (Ollama/OpenAI)   │                │
│                                    └─────────┬─────────┘                │
│                                              │                          │
│                    Action: {                 │                          │
│                      action: "twitch.clip",  │                          │
│                      params: {duration: 30}, │                          │
│                      reply: "Creando clip"   │                          │
│                    }                         │                          │
│                                              ▼                          │
│                                    ┌───────────────────┐                │
│                                    │      Brain        │                │
│                                    │   (Dispatcher)    │                │
│                                    └─────────┬─────────┘                │
│                                              │                          │
│                         ┌────────────────────┼────────────────────┐     │
│                         ▼                    ▼                    ▼     │
│                   ┌──────────┐         ┌──────────┐         ┌──────────┐│
│                   │ Twitch   │         │   OBS    │         │  Music   ││
│                   │ Executor │         │ Executor │         │ Executor ││
│                   └────┬─────┘         └──────────┘         └──────────┘│
│                        │                                                │
│                        ▼                                                │
│              [POST Twitch API]                                          │
│                        │                                                │
│                        ▼                                                │
│                   Result: OK                                            │
│                        │                                                │
│                        ▼                                                │
│                   ┌───────────────────┐                                 │
│                   │   TTS Provider    │                                 │
│                   │  (Piper/OpenAI)   │                                 │
│                   └─────────┬─────────┘                                 │
│                             │                                           │
│                             ▼                                           │
│                   🔊 "Creando clip de 30 segundos"                      │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## ⚙️ CONFIGURACIÓN (ana.config.yaml)

```yaml
general:
  language: "es"
  log_level: "info"        # debug, info, warn, error

stt:
  provider: "whisper"      # whisper | openai
  whisper:
    binary_path: "./bin/whisper"
    model_path: "./assets/models/whisper/ggml-base.bin"
    language: "es"

llm:
  provider: "ollama"       # ollama | openai
  ollama:
    url: "http://localhost:11434"
    model: "llama3.2:3b"
    timeout_seconds: 30

tts:
  provider: "piper"        # piper | openai
  piper:
    binary_path: "./bin/piper"
    model_path: "./assets/voices/piper/es_ES-davefx-medium.onnx"

twitch:
  enabled: true
  client_id: "..."
  client_secret: "..."
  broadcaster_id: "..."

obs:
  enabled: true
  url: "ws://localhost:4455"
  password: "..."

music:
  enabled: true
  folders: ["./music"]
  default_volume: 0.5
```

---

## 📦 DEPENDENCIAS PRINCIPALES

```go
require (
    github.com/go-resty/resty/v2      // HTTP client
    github.com/gordonklaus/portaudio   // Audio capture [PENDIENTE]
    github.com/gorilla/websocket       // OBS WebSocket
    github.com/rs/zerolog              // Logging
    github.com/spf13/viper             // Configuration
    golang.design/x/hotkey             // Hotkeys [PENDIENTE]
)
```

---

## 🚧 MÓDULOS PENDIENTES

### 1. Audio Capture (internal/audio/)
- Captura continua de micrófono con PortAudio
- Ring buffer para audio
- VAD (Voice Activity Detection)
- Wake word detection ("Ana")

### 2. Hotkey System (internal/hotkey/)
- Hotkeys globales del sistema
- Modo push-to-talk (hold/toggle)
- Soporte multiplataforma (Windows, Linux, macOS)

### 3. UI (internal/ui/)
- System tray icon
- Notificaciones del sistema
- Log viewer

### 4. Twitch OAuth (internal/executor/twitch/auth.go)
- Flujo OAuth PKCE completo
- Servidor HTTP local para callback
- Refresh automático de tokens

---

## 🧪 CÓMO PROBAR

```bash
# Modo interactivo (texto)
./ana

# Probar un comando específico
./ana -test -command "crea un clip de 30 segundos"

# Ver estado
Ana> status

# Comandos de ejemplo
Ana> crea un clip
Ana> cambia a la escena gameplay
Ana> pon música rock
Ana> banea a troll123
```

---

## 📝 CONVENCIONES DE CÓDIGO

1. **Nombres de archivos**: snake_case (ej: `openai_stt.go`)
2. **Interfaces**: Terminan en `er` o son descriptivas (ej: `Provider`, `Executor`)
3. **Constructores**: `New...()` (ej: `NewOllamaProvider()`)
4. **Errores**: Siempre wrap con contexto (ej: `fmt.Errorf("failed to X: %w", err)`)
5. **Logging**: Usar `logger.Component("nombre")` para cada módulo
6. **Context**: Siempre pasar `context.Context` como primer parámetro

---

## 🔗 INTEGRACIONES EXTERNAS

### Twitch Helix API
- Base URL: `https://api.twitch.tv/helix`
- Auth: OAuth Bearer token + Client-ID header
- Endpoints: `/clips`, `/channels`, `/moderation/bans`, `/users`

### OBS WebSocket 5.x
- URL: `ws://localhost:4455`
- Auth: SHA256 challenge-response
- Requests: `SetCurrentProgramScene`, `SetInputVolume`, `SetInputMute`, etc.

### Ollama
- URL: `http://localhost:11434`
- Endpoint: `POST /api/generate`
- Format: `"format": "json"` para forzar JSON output

### Whisper.cpp
- Binario: `./bin/whisper`
- Args: `--model`, `--file`, `--language`, `--output-txt`
- Input: WAV 16kHz mono 16-bit

### Piper TTS
- Binario: `./bin/piper`
- Args: `--model`, `--output_file`
- Input: texto via stdin
- Output: WAV 22050Hz mono

---

## 🎯 PRÓXIMOS PASOS SUGERIDOS

1. [ ] Implementar `internal/audio/` con PortAudio
2. [ ] Implementar `internal/hotkey/` con golang.design/x/hotkey
3. [ ] Agregar wake word detection (Porcupine o propio)
4. [ ] Implementar OAuth flow completo para Twitch
5. [ ] Agregar UI con system tray (systray library)
6. [ ] Tests unitarios para cada módulo
7. [ ] Documentación de API interna
8. [ ] Instalador para Windows/Linux/macOS

---

# FIN DEL CONTEXTO
# Usa este archivo como referencia para entender y extender AnaStreamer
