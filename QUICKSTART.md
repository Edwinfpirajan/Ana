# ⚡ Quick Start - Ana Streamer

Guía rápida para comenzar en 5 minutos.

## 1️⃣ Requisitos Mínimos

Verifica que tengas instalado:

```bash
# Go 1.22+
go version

# GCC
gcc --version

# PortAudio
pkg-config --modversion portaudio-2.0
```

Si algo falta, ve a [BUILDING.md](BUILDING.md) para instrucciones de instalación.

---

## 2️⃣ Compilar Ana

### Windows
```bash
./build.bat
```

### Linux/macOS
```bash
./build.sh
```

Espera a que finalice. Deberías ver:
```
============================================================================
 BUILD SUCCESSFUL
============================================================================
```

---

## 3️⃣ Configurar Ana

```bash
# Copiar configuración de ejemplo
cp config/ana.config.example.yaml config/ana.config.yaml

# Editar configuración (opcional, ya tiene valores por defecto)
nano config/ana.config.yaml
```

**Cambios importantes (si lo necesitas):**
- `general.streamer_name`: Tu nombre (por defecto: "Ferchando")
- `audio.device`: "default" funciona para la mayoría
- `stt.provider`: "whisper" (local) o "openai" (cloud)
- `llm.provider`: "ollama" (local) o "openai" (cloud)
- `tts.provider`: "piper" (local) o "openai" (cloud)

---

## 4️⃣ Ejecutar Ana

### Windows
```bash
./ana.exe
```

### Linux/macOS
```bash
./ana
```

Deberías ver:
```
🎤 Ana Streamer Active
════════════════════════════════════════════════════════

🔊 How to use:
   1. Say 'Ana' to activate persistent session
   2. Keep talking - no need to repeat 'Ana'
   3. Say 'Adiós Ana' or similar to deactivate

⌨️  Hotkey: F4 (press and hold)

Press Ctrl+C to exit
════════════════════════════════════════════════════════
```

---

## 5️⃣ Probar Ana

Abre otra terminal en la carpeta del proyecto:

```bash
# Modo texto (sin audio)
echo "Ana, dame el status" | ./ana
```

O por voz (si tienes Whisper y Ollama corriendo):
- Di "Ana" para activar
- Di algo como "Crea un clip" o "que hora es"

---

## 🐛 Problemas Comunes

### "PortAudio is required"
Compilaste sin PortAudio. Usa `./build.bat` o `./build.sh` en lugar de `go build`.

### "Ollama not available"
Ollama no está corriendo. Abre otra terminal:
```bash
ollama serve
```

### "Whisper binary not found"
Descarga Whisper.cpp en `./bin/whisper/`. Ver [BUILDING.md](BUILDING.md).

---

## 📚 Documentación Completa

Para más detalles, ve a:
- [README.md](README.md) - Documentación general
- [BUILDING.md](BUILDING.md) - Guía de compilación detallada
- [config/ana.config.example.yaml](config/ana.config.example.yaml) - Todas las opciones

---

## 🎯 Próximos Pasos

1. Descarga modelos de Whisper y Piper
2. Configura Ollama con un modelo
3. (Opcional) Integra con Twitch/OBS

¡Listo! Ana Streamer está corriendo. 🚀
