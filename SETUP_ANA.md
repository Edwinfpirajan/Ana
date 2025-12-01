# ANA Streamer - Configuración

Ana Streamer es un asistente de voz inteligente para streamers, construido 100% en Go.

## Clonación y Setup

```bash
git clone https://github.com/Edwinfpirajan/Ana.git
cd Ana
```

## Descargar Modelos y Voces (Importante)

Los modelos de Whisper y las voces de Piper NO están en el repositorio por su tamaño (>300MB total).

### Opción 1: Descargar desde la rama main

```bash
git checkout main
# Copiar archivos de assets/models y assets/voices
git checkout ana
```

### Opción 2: Descargar manualmente

1. **Modelo Whisper** (141 MB):
   - Descargar: https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin
   - Guardar en: `assets/models/whisper/ggml-base.bin`

2. **Voces Piper** (60-108 MB cada una):
   - Descargar desde: https://huggingface.co/rhasspy/piper/tree/main/voices
   - Guardar en: `assets/voices/piper/[voice-name]/`

## Compilación

Con MSYS2 y GCC instalados:

```bash
bash -c "export CGO_ENABLED=1 CC=/c/msys64/mingw64/bin/gcc && \
         go build -tags portaudio -o ana.exe ./cmd/ana/main.go"
```

O usar el script:

```bash
./build.bat
```

## Ejecución

```bash
# Opción 1: Directamente
.\ana.exe

# Opción 2: Ventana que permanece abierta
cmd /k ".\ana.exe"

# Opción 3: Usar script
.\run_debug.bat
```

## Configuración

Editar `config/ana.config.yaml`:
- STT: Whisper local (recomendado) o OpenAI
- LLM: Ollama local (recomendado) o OpenAI
- TTS: Piper local (recomendado) o OpenAI
- Wake word: "ana"

## Requisitos

- Go 1.20+
- MSYS2 con GCC (para compilación)
- PortAudio (para audio capture)
- Ollama (para LLM local)
- Piper (para TTS local)

Ver README_ANA.md para uso completo.
