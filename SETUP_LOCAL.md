# 🎙️ Guía de Setup Local - Ana 100% Sin IA en la Nube

Esta guía te ayudará a configurar Ana para que funcione **completamente local** sin depender de ningún servicio de IA en la nube (OpenAI, etc.).

---

## 🎯 ¿Qué significa "100% Local"?

- ✅ **STT (Speech-to-Text)**: Whisper.cpp ejecutándose en tu PC
- ✅ **LLM (Language Model)**: Ollama con modelos locales
- ✅ **TTS (Text-to-Speech)**: Piper generando voz en tu PC
- ✅ **Sin conexión a internet requerida** para funcionalidades principales
- ✅ **Sin API keys** de servicios externos
- ✅ **Privacidad total**: Ningún dato sale de tu computadora

---

## 🚀 Setup Rápido (Recomendado)

### Opción A: Script Automático

```powershell
# Ejecutar script maestro de instalación
.\scripts\setup_local.ps1
```

Este script instalará automáticamente:
- Piper TTS + voz en español
- Modelos de Whisper.cpp
- Verificará Ollama
- Configurará Ana para modo local

### Opción B: Instalación Manual Paso a Paso

Sigue las secciones a continuación.

---

## 📦 Requisitos Previos

- **Windows 10/11** (x64)
- **Go 1.22+** instalado
- **PowerShell 5.1+**
- **~5 GB de espacio en disco** (para modelos)
- **8 GB+ RAM recomendado**

---

## 🔧 Instalación Componente por Componente

### 1️⃣ Ollama (LLM Local)

**¿Qué hace?**: Procesa tus comandos de voz y decide qué acción ejecutar.

#### Instalación:

```powershell
# Opción A: Descarga desde la web
# Visita: https://ollama.ai/download
# Descarga el instalador para Windows y ejecútalo

# Opción B: Con Winget (Windows Package Manager)
winget install Ollama.Ollama
```

#### Configuración:

```powershell
# 1. Iniciar servidor Ollama (déjalo corriendo en segundo plano)
ollama serve

# 2. En otra terminal, descargar modelo (3 GB aproximadamente)
ollama pull llama3.2:3b

# Verificar instalación
ollama list
```

**Modelos alternativos**:
- `llama3.2:1b` - Más rápido, menos preciso (1 GB)
- `mistral:7b` - Más potente, más lento (4 GB)
- `phi3:mini` - Balance intermedio (2 GB)

---

### 2️⃣ Piper (TTS Local)

**¿Qué hace?**: Convierte texto a voz para que Ana "hable".

#### Instalación Automática:

```powershell
# Instalar Piper
.\scripts\install_piper.ps1

# Descargar voz en español (España)
.\scripts\download_voices.ps1

# O voz de México
.\scripts\download_voices.ps1 -Language es_MX -Voice ald -Quality medium
```

#### Instalación Manual:

```powershell
# 1. Crear directorio
mkdir -p bin\piper

# 2. Descargar desde GitHub
# Visita: https://github.com/rhasspy/piper/releases
# Descarga: piper_windows_amd64.zip
# Extrae el contenido en: bin\piper\

# 3. Descargar modelo de voz
mkdir -p assets\voices\piper

# URL del modelo (copia en navegador):
# https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx
# Guardar en: assets\voices\piper\es_ES-davefx-medium.onnx

# También descargar el archivo JSON:
# https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx.json
# Guardar en: assets\voices\piper\es_ES-davefx-medium.onnx.json
```

#### Probar Piper:

```powershell
echo "Hola, soy Ana" | .\bin\piper\piper.exe --model .\assets\voices\piper\es_ES-davefx-medium.onnx --output_file test.wav
```

---

### 3️⃣ Whisper.cpp (STT Local)

**¿Qué hace?**: Convierte tu voz en texto.

#### Instalación Automática (Solo Modelos):

```powershell
# Descargar modelo base (142 MB)
.\scripts\install_whisper.ps1 -Model base

# O modelo pequeño para mejor precisión (466 MB)
.\scripts\install_whisper.ps1 -Model small
```

#### Instalación del Binario:

**Opción A - Descarga Precompilada (Recomendada)**:

1. Visita: [Whisper.cpp Releases](https://github.com/ggerganov/whisper.cpp/releases)
2. Descarga: `whisper-bin-x64.zip` (Windows)
3. Extrae `main.exe` en: `bin\whisper\`

**Opción B - Compilar desde Fuente (Avanzado)**:

```powershell
# Requiere: Visual Studio 2022 + CMake + Git

git clone https://github.com/ggerganov/whisper.cpp.git
cd whisper.cpp
mkdir build
cd build
cmake ..
cmake --build . --config Release

# Copiar ejecutable
copy bin\Release\main.exe ..\..\..\bin\whisper\main.exe
```

#### Probar Whisper:

```powershell
# Grabar un audio WAV y probarlo
.\bin\whisper\main.exe -m .\assets\models\whisper\ggml-base.bin -f test.wav
```

---

## ⚙️ Configuración de Ana

### Configuración Ya Aplicada

Tu archivo [config/ana.config.yaml](config/ana.config.yaml) ya está configurado para modo local:

```yaml
stt:
  provider: "whisper"  # ✅ Local
  whisper:
    binary_path: "./bin/whisper/main.exe"
    model_path: "./assets/models/whisper/ggml-base.bin"
    language: "es"

llm:
  provider: "ollama"   # ✅ Local
  ollama:
    url: "http://localhost:11434"
    model: "llama3.2:3b"

tts:
  provider: "piper"    # ✅ Local
  piper:
    binary_path: "./bin/piper/piper.exe"
    model_path: "./assets/voices/piper/es_ES-davefx-medium.onnx"
    speed: 1.0
```

### Configuración Híbrida (Opcional)

Si quieres **fallback** a OpenAI cuando los servicios locales fallen:

```yaml
stt:
  provider: "whisper"  # Solo local

llm:
  provider: "auto"     # Intenta Ollama → OpenAI

tts:
  provider: "auto"     # Intenta Piper → OpenAI
```

---

## 🏃 Ejecutar Ana

### 1. Iniciar Ollama (si no está corriendo)

```powershell
# En una terminal separada (déjala abierta)
ollama serve
```

### 2. Compilar Ana

```powershell
# Desde la raíz del proyecto
go build -o ana.exe ./cmd/ana
```

### 3. Ejecutar Ana

```powershell
# Modo interactivo (texto)
.\ana.exe

# Modo test
.\ana.exe -test -command "crea un clip"

# Modo voz (requiere audio configurado)
.\ana.exe -voice
```

---

## 🧪 Verificar que Todo Funciona

### Checklist Pre-vuelo:

```powershell
# 1. Verificar Ollama
ollama list
# Debe mostrar: llama3.2:3b

# 2. Verificar Ollama está corriendo
Invoke-WebRequest http://localhost:11434/api/version
# Debe responder con versión

# 3. Verificar Piper
.\bin\piper\piper.exe --version
# Debe mostrar versión de Piper

# 4. Verificar voz de Piper
dir assets\voices\piper\*.onnx
# Debe listar: es_ES-davefx-medium.onnx

# 5. Verificar Whisper (si compilaste)
.\bin\whisper\main.exe --help
# Debe mostrar opciones

# 6. Verificar modelos Whisper
dir assets\models\whisper\*.bin
# Debe listar: ggml-base.bin
```

---

## 📊 Comparativa: Local vs OpenAI

| Aspecto | Local | OpenAI |
|---------|-------|--------|
| **Costo** | Gratis | Pago por uso |
| **Privacidad** | Total | Envía datos a la nube |
| **Velocidad** | Depende de tu PC | Rápido (servidores potentes) |
| **Conexión** | No requerida | Requiere internet |
| **Calidad STT** | Muy buena | Excelente |
| **Calidad TTS** | Buena | Excelente |
| **Calidad LLM** | Buena (3B params) | Excelente (70B+ params) |
| **Setup** | Complejo | Trivial (solo API key) |

---

## 🐛 Solución de Problemas

### Problema: "Ollama not running"

```powershell
# Solución: Iniciar Ollama
ollama serve
```

### Problema: "Piper binary not found"

```powershell
# Verificar ruta
dir bin\piper\piper.exe

# Re-ejecutar instalación
.\scripts\install_piper.ps1
```

### Problema: "Whisper model not found"

```powershell
# Descargar modelo
.\scripts\install_whisper.ps1 -Model base
```

### Problema: "STT failed: file not found"

**Causa**: Whisper.cpp no instalado (requiere compilación manual)

**Solución Temporal**: Usa OpenAI STT mientras tanto

```yaml
stt:
  provider: "openai"
  openai:
    api_key: "${OPENAI_API_KEY}"
```

### Problema: "LLM timeout"

**Causa**: Modelo Ollama muy grande para tu PC

**Solución**: Usar modelo más pequeño

```powershell
# Descargar modelo más ligero
ollama pull llama3.2:1b

# Actualizar config
# llm.ollama.model: "llama3.2:1b"
```

---

## 🎯 Optimizaciones de Rendimiento

### Para PCs de gama baja:

```yaml
llm:
  ollama:
    model: "llama3.2:1b"  # Modelo más pequeño

stt:
  whisper:
    model_path: "./assets/models/whisper/ggml-tiny.bin"  # Modelo más rápido

tts:
  piper:
    model_path: "./assets/voices/piper/es_ES-sharvard-low.onnx"  # Voz de baja calidad pero rápida
```

### Para PCs potentes:

```yaml
llm:
  ollama:
    model: "mistral:7b"  # Modelo más inteligente

stt:
  whisper:
    model_path: "./assets/models/whisper/ggml-medium.bin"  # Mejor precisión

tts:
  piper:
    model_path: "./assets/voices/piper/es_ES-davefx-high.onnx"  # Mejor calidad de voz
```

---

## 📚 Recursos Adicionales

- [Ollama Models Library](https://ollama.ai/library)
- [Piper Voice Samples](https://rhasspy.github.io/piper-samples/)
- [Whisper.cpp GitHub](https://github.com/ggerganov/whisper.cpp)
- [Ana README](README.md)

---

## 🤝 ¿Necesitas Ayuda?

- **Issues**: Abre un issue en GitHub
- **Discord**: [Enlace al servidor] (si existe)
- **Email**: [tu email]

---

**¡Disfruta de tu asistente de voz 100% local!** 🎙️🚀
