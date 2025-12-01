# 📜 Scripts de Instalación - Ana Local

Este directorio contiene scripts PowerShell para configurar Ana en modo 100% local.

---

## 🎯 Scripts Disponibles

### 1. `setup_local.ps1` - Script Maestro (Recomendado)

**Descripción**: Instalador completo que ejecuta todos los pasos automáticamente.

**Uso**:
```powershell
# Setup completo por defecto
.\setup_local.ps1

# Omitir componentes específicos
.\setup_local.ps1 -SkipPiper
.\setup_local.ps1 -SkipWhisper
.\setup_local.ps1 -SkipOllama

# Personalizar modelos
.\setup_local.ps1 -WhisperModel small -PiperVoice es_MX-ald-medium
```

**Parámetros**:
- `-SkipPiper`: No instalar Piper TTS
- `-SkipWhisper`: No instalar Whisper.cpp
- `-SkipOllama`: No verificar Ollama
- `-WhisperModel`: Modelo Whisper (tiny, base, small, medium, large) - Default: base
- `-PiperVoice`: Voz Piper - Default: es_ES-davefx-medium

**Acciones**:
1. Verifica/instala Ollama
2. Descarga Piper + voz española
3. Descarga modelos Whisper
4. Verifica configuración
5. Muestra resumen del estado

---

### 2. `install_piper.ps1` - Instalador de Piper TTS

**Descripción**: Descarga e instala Piper para síntesis de voz local.

**Uso**:
```powershell
# Instalación por defecto (Windows x64)
.\install_piper.ps1

# Especificar versión
.\install_piper.ps1 -Version "2023.11.14-2"

# Especificar arquitectura
.\install_piper.ps1 -Arch arm64
```

**Parámetros**:
- `-Version`: Versión de Piper - Default: 2023.11.14-2
- `-Arch`: Arquitectura (amd64, arm64) - Default: amd64

**Salida**:
- Binario instalado en: `bin\piper\piper.exe`
- Muestra versión instalada
- Sugiere siguiente paso (descargar voces)

---

### 3. `download_voices.ps1` - Descargador de Voces Piper

**Descripción**: Descarga modelos de voz en español para Piper.

**Uso**:
```powershell
# Voz por defecto (España, masculina, calidad media)
.\download_voices.ps1

# Voz de México
.\download_voices.ps1 -Language es_MX -Voice ald -Quality medium

# Voz de alta calidad
.\download_voices.ps1 -Language es_ES -Voice davefx -Quality high

# Voz rápida (baja calidad)
.\download_voices.ps1 -Language es_ES -Voice sharvard -Quality low
```

**Parámetros**:
- `-Language`: Código de idioma (es_ES, es_MX) - Default: es_ES
- `-Voice`: Nombre de voz (davefx, ald, mls, etc.) - Default: davefx
- `-Quality`: Calidad (low, medium, high) - Default: medium

**Salida**:
- Modelo ONNX en: `assets\voices\piper\{nombre}.onnx`
- Archivo JSON de configuración
- Opción de probar la voz al finalizar

**Voces disponibles en español**:
- `es_ES-davefx-medium` (España, masculina) ⭐ Recomendada
- `es_ES-mls-medium` (España, múltiples hablantes)
- `es_MX-ald-medium` (México, masculina)

---

### 4. `install_whisper.ps1` - Instalador de Whisper.cpp

**Descripción**: Descarga modelos Whisper y guía instalación del binario.

**Uso**:
```powershell
# Modelo base (recomendado - 142 MB)
.\install_whisper.ps1

# Modelo pequeño (mejor precisión - 466 MB)
.\install_whisper.ps1 -Model small

# Modelo tiny (más rápido - 75 MB)
.\install_whisper.ps1 -Model tiny

# Modelo medium (muy preciso - 1.5 GB)
.\install_whisper.ps1 -Model medium
```

**Parámetros**:
- `-Model`: Modelo Whisper (tiny, base, small, medium, large) - Default: base
- `-Language`: Idioma - Default: es

**Salida**:
- Modelo descargado en: `assets\models\whisper\ggml-{model}.bin`
- Instrucciones para descargar binario precompilado
- Sugerencias de configuración

**Tamaños de modelos**:
- `tiny` = 75 MB (rápido, menos preciso)
- `base` = 142 MB (balance recomendado) ⭐
- `small` = 466 MB (más preciso)
- `medium` = 1.5 GB (muy preciso, lento)
- `large` = 2.9 GB (máxima precisión)

**IMPORTANTE**: Este script solo descarga modelos. El binario `main.exe` debe descargarse manualmente de:
https://github.com/ggerganov/whisper.cpp/releases

---

## 🔄 Flujo de Trabajo Recomendado

### Setup Inicial Completo:

```powershell
# 1. Ejecutar script maestro
.\scripts\setup_local.ps1

# 2. Instalar Ollama manualmente (si no está)
# Visitar: https://ollama.ai/download

# 3. Iniciar Ollama
ollama serve

# 4. Descargar modelo LLM
ollama pull llama3.2:3b

# 5. Descargar Whisper binario manualmente
# https://github.com/ggerganov/whisper.cpp/releases

# 6. Compilar y ejecutar Ana
go build -o ana.exe ./cmd/ana
.\ana.exe
```

### Setup Parcial (Solo TTS):

```powershell
# Solo necesitas que Ana hable?
.\scripts\install_piper.ps1
.\scripts\download_voices.ps1

# Configurar en ana.config.yaml:
# tts.provider: "piper"
```

### Actualizar Componentes:

```powershell
# Actualizar Piper a nueva versión
.\scripts\install_piper.ps1 -Version "2024.01.15-1"

# Descargar modelo Whisper más grande
.\scripts\install_whisper.ps1 -Model small

# Agregar más voces
.\scripts\download_voices.ps1 -Language es_MX -Voice ald
```

---

## 📂 Estructura de Directorios Creada

Después de ejecutar los scripts, tendrás:

```
ana/
├── bin/
│   ├── piper/
│   │   └── piper.exe              # Binario Piper TTS
│   └── whisper/
│       └── main.exe               # Binario Whisper (manual)
├── assets/
│   ├── voices/
│   │   └── piper/
│   │       ├── es_ES-davefx-medium.onnx
│   │       └── es_ES-davefx-medium.onnx.json
│   └── models/
│       └── whisper/
│           └── ggml-base.bin      # Modelo Whisper
└── config/
    └── ana.config.yaml         # Configuración
```

---

## ⚙️ Variables de Entorno

Los scripts no requieren variables de entorno, pero Ana sí:

```powershell
# Crear archivo .env (opcional para fallback OpenAI)
echo "OPENAI_API_KEY=sk-..." > .env
echo "ANA_ENV=development" >> .env
```

---

## 🐛 Troubleshooting

### Error: "No se puede ejecutar scripts"

```powershell
# Solución: Habilitar ejecución de scripts
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Error: "Invoke-WebRequest falla"

```powershell
# Solución: Usar TLS 1.2
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
```

### Error: "Access Denied al descargar"

```powershell
# Solución: Ejecutar PowerShell como Administrador
# O descargar manualmente desde el navegador
```

---

## 🔗 Enlaces Útiles

- [Piper GitHub](https://github.com/rhasspy/piper)
- [Piper Voices](https://huggingface.co/rhasspy/piper-voices)
- [Whisper.cpp GitHub](https://github.com/ggerganov/whisper.cpp)
- [Ollama Download](https://ollama.ai/download)
- [Ollama Models](https://ollama.ai/library)

---

## 📝 Notas

- Todos los scripts están diseñados para **Windows PowerShell**
- Los scripts son **idempotentes**: puedes ejecutarlos múltiples veces
- Los scripts verifican si los componentes ya existen antes de descargar
- Los archivos descargados se cachean, no se re-descargan si ya existen

---

**¿Problemas?** Abre un issue en GitHub con el output completo del script.
