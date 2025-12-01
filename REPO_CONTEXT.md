# AnaStreamer - Resumen completo

AnaStreamer es un asistente de voz local para streamers escrito en Go 1.22+. El pipeline general es: captura de audio → STT (Whisper.cpp u OpenAI) → LLM (Ollama u OpenAI) → `llm.Action` → ejecutores (Twitch, OBS, música) → TTS (Piper u OpenAI). Todo se configura mediante `config/ana.config.yaml` y los scripts `build.*`, `run.*` y los del directorio `scripts/`.

## Arquitectura general

```
Audio Capture → STT → Pipeline → Brain → Executor → (TTS | Log)
```

- `cmd/ana/main.go` arranca: logger, carga config, inicializa proveedores y el pipeline.
- `internal/audio/` captura audio (PortAudio), aplica VAD y wake word “Ana”.
- `internal/stt/` contiene Whisper local y cliente OpenAI (ambos exponen `stt.Provider`).
- `internal/pipeline/pipeline.go` filtra transcripciones sin “Ana”, llama al `brain` y dispara callbacks.
- `internal/brain/brain.go` manda el texto al LLM configurado y envía respuestas al TTS si hay.
- `internal/llm/` incluye prompts (`prompt.go`), cliente Ollama, cliente OpenAI y el struct `llm.Action`.
- `llm.Action` tiene `action`, `params` y `reply`. Siempre se espera un JSON válido.
- `internal/executor/` agrupa ejecutores para Twitch, OBS y música local; todos siguen `executor.Executor`.
- `internal/tts/` gestiona Piper local y OpenAI TTS, ambos implementan `tts.Provider`.
- `pkg/logger` envuelve zerolog; `pkg/utils` incluye helpers de audio/JSON/procesos.

## Acciones soportadas

| Dominio | Acciones |
|---------|----------|
| `twitch.*` | `clip`, `title`, `category`, `ban`, `timeout`, `unban` |
| `obs.*` | `scene`, `source.show`, `source.hide`, `volume`, `mute`, `unmute`, `text` |
| `music.*` | `play`, `pause`, `resume`, `next`, `previous`, `volume`, `stop` |
| `system.*` | `status`, `help`, `none` |

La respuesta del LLM debe ser JSON y decir qué acción ejecutar. `system.none` se usa para conversaciones sin efecto.

## Control de plataformas

- **Twitch:** `internal/executor/twitch/client.go` usa Helix con OAuth. Ejecuta clips, títulos/categorías y moderación.
- **OBS:** `internal/executor/obs/client.go` se conecta a OBS WebSocket 5.x y permite escenas, fuentes, volumen y texto.
- **Música local:** `internal/executor/music/player.go` explora carpetas (`music.folders`), construye playlists, y usa `mpv`/`ffplay`/`afplay`. Soporta play/pause/resume/next/prev/volume/stop.
- **Integrar Spotify:** se puede añadir un executor adicional (`music.spotify`) que implemente `executor.Executor`, consuma la API Web y traduzca comandos como `music.spotify.play`, `music.spotify.pause`, `music.spotify.volume` para el streaming remoto.

## Configuración relevante

- Copia `config/ana.config.example.yaml` a `config/ana.config.yaml` y ajusta STT, LLM, TTS, Twitch, OBS, música y la palabra de activación “Ana”.
- `internal/config/loader.go` busca `ana.config` en `./config`, `~/.ana` y `/etc/ana`, usa prefijo `ANA_` para env vars y aplica defaults.
- `internal/config/defaults.go` define el wake word `ana`, rutas predeterminadas y valores de proveedores.
- Hay scripts en `scripts/` para setup completo (`setup_local.ps1`, instaladores, etc.).
- Los recursos `assets/` contienen modelos y sonidos; `bin/` guarda binarios de terceros (Whisper, Piper).

## Comandos comunes

```bash
# Build con PortAudio
# Windows
./build.bat

# Linux/macOS
./build.sh

# Ejecutar
./ana

# Testeo rápido
./ana -test -command "crea un clip"

# Tests Go
go test ./...
```

## Consideraciones adicionales

- Puedes mezclar proveedores configurando `llm.provider: auto` y `tts.provider: auto` para verificar Ollama/Piper primero y caer en OpenAI si hace falta.
- Los activadores de wake word y hotkey (`internal/hotkey/`) aún están pendientes.
- Documentación en AGENTS, AI_PROMPT, BUILDING, QUICKSTART y scripts describe la operación de Ana y las reglas para agentes.
- Para permitir música en Spotify basta con construir un executor nuevo que use OAuth y la API Web; se comunica igual con el brain porque sigue la interfaz.

Este archivo resume cómo el código, las acciones y las integraciones trabajan juntos. Si quieres que lo replique en la copia antigua también o que lo enlace desde un README, dímelo.
