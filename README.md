# 🎙️ AnaStreamer

**Tu asistente de voz personal para streaming** - 100% local, sin necesidad de servidores externos.

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## ✨ Características

- 🎤 **Always-Listening** con wake word "Ana"
- ⌨️ **Push-to-Talk** con hotkey configurable
- 🗣️ **STT Local** con Whisper.cpp (o OpenAI como alternativa)
- 🧠 **LLM Local** con Ollama (o OpenAI como alternativa)  
- 🔊 **TTS Local** con Piper (o OpenAI como alternativa)
- 📺 **Control de Twitch**: clips, título, categoría, bans
- 🎬 **Control de OBS**: escenas, fuentes, volumen
- 🎵 **Reproductor de música** integrado

## 🚀 Inicio Rápido

### Requisitos Previos

1. **Go 1.22+** instalado
2. **Ollama** corriendo localmente (para LLM)
3. **Whisper.cpp** compilado (para STT)
4. **Piper** instalado (para TTS)

### Instalación

```bash
# Clonar repositorio
git clone https://github.com/tuusuario/ana-streamer.git
cd ana-streamer

# Descargar dependencias
go mod download

# Copiar configuración de ejemplo
cp config/ana.config.example.yaml config/ana.config.yaml

# Editar configuración
nano config/ana.config.yaml

# Compilar
go build -o ana ./cmd/ana

# Ejecutar
./ana
```

### Descargar Modelos

```bash
# Descargar modelo Whisper
mkdir -p assets/models/whisper
wget https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin \
  -O assets/models/whisper/ggml-base.bin

# Descargar voz Piper (español)
mkdir -p assets/voices/piper
wget https://github.com/rhasspy/piper/releases/download/v1.2.0/voice-es_ES-davefx-medium.tar.gz
tar -xzf voice-es_ES-davefx-medium.tar.gz -C assets/voices/piper/

# Instalar modelo en Ollama
ollama pull llama3.2:3b
```

## 📖 Uso

### Modo Interactivo (Texto)

```bash
./ana
```

Escribe comandos directamente:
```
Ana> crea un clip
Ana> cambia a la escena gameplay
Ana> pon música
Ana> status
Ana> quit
```

### Modo Test

```bash
./ana -test -command "crea un clip de 30 segundos"
```

## 🎯 Comandos Disponibles

### Twitch
| Comando | Ejemplo |
|---------|---------|
| Crear clip | "Ana, crea un clip" |
| Cambiar título | "Cambia el título a Jugando Minecraft" |
| Cambiar categoría | "Pon la categoría Just Chatting" |
| Banear usuario | "Banea a troll123" |
| Timeout | "Dale timeout de 5 minutos a spammer" |

### OBS
| Comando | Ejemplo |
|---------|---------|
| Cambiar escena | "Cambia a la escena de inicio" |
| Mostrar fuente | "Muestra la webcam" |
| Ocultar fuente | "Oculta el chat" |
| Cambiar volumen | "Sube el volumen del micrófono" |
| Mutear | "Mutea el audio del escritorio" |

### Música
| Comando | Ejemplo |
|---------|---------|
| Reproducir | "Pon música" |
| Pausar | "Pausa la música" |
| Siguiente | "Siguiente canción" |
| Volumen | "Baja el volumen de la música" |

## ⚙️ Configuración

Edita `config/ana.config.yaml`:

```yaml
# Seleccionar proveedores
stt:
  provider: "whisper"  # whisper | openai

llm:
  provider: "ollama"   # ollama | openai

tts:
  provider: "piper"    # piper | openai

# Configurar Twitch
twitch:
  enabled: true
  client_id: "tu_client_id"
  client_secret: "tu_client_secret"
  broadcaster_id: "tu_broadcaster_id"

# Configurar OBS
obs:
  enabled: true
  url: "ws://localhost:4455"
  password: "tu_password"

# Configurar música
music:
  enabled: true
  folders:
    - "./music"
    - "D:/Music/Stream"
```

## 🏗️ Arquitectura

```
┌─────────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│   Audio     │───▶│   STT   │───▶│   LLM   │───▶│  Brain  │
│  Capture    │    │ Whisper │    │ Ollama  │    │         │
└─────────────┘    └─────────┘    └─────────┘    └────┬────┘
                                                      │
                                              ┌───────┴───────┐
                                              ▼       ▼       ▼
                                          ┌──────┐┌──────┐┌──────┐
                                          │Twitch││ OBS  ││Music │
                                          └──────┘└──────┘└──────┘
```

## 📁 Estructura del Proyecto

```
AnaStreamer/
├── cmd/ana/          # Punto de entrada
├── internal/
│   ├── config/          # Configuración
│   ├── audio/           # Captura de audio
│   ├── stt/             # Speech-to-Text
│   ├── llm/             # Language Model
│   ├── tts/             # Text-to-Speech
│   ├── brain/           # Orquestador
│   ├── executor/        # Ejecutores de acciones
│   │   ├── twitch/
│   │   ├── obs/
│   │   └── music/
│   └── pipeline/        # Pipeline de procesamiento
├── pkg/
│   ├── logger/          # Logging
│   └── utils/           # Utilidades
├── config/              # Archivos de configuración
├── assets/              # Modelos y recursos
└── docs/                # Documentación
```

## 🔧 Desarrollo

```bash
# Ejecutar tests
go test ./...

# Compilar para producción
go build -ldflags "-s -w" -o ana ./cmd/ana

# Cross-compile para Windows
GOOS=windows GOARCH=amd64 go build -o ana.exe ./cmd/ana
```

## 📝 Licencia

MIT License - ver [LICENSE](LICENSE)

## 🤝 Contribuir

1. Fork el proyecto
2. Crea tu rama (`git checkout -b feature/nueva-caracteristica`)
3. Commit tus cambios (`git commit -am 'Añade nueva característica'`)
4. Push a la rama (`git push origin feature/nueva-caracteristica`)
5. Abre un Pull Request

---

**¡Hecho con ❤️ para streamers!**
