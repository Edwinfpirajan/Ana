# 🚀 Cómo Ejecutar Ana Streamer

## ⚠️ IMPORTANTE: Usa los scripts de build, NO `go run`

### ❌ ESTO NO FUNCIONA
```bash
go run cmd/ana/main.go
# Error: "PortAudio build tag is required for audio capture"
```

### ✅ ESTO SÍ FUNCIONA

**En Windows (desde PowerShell o CMD):**
```bash
./build.bat
./ana.exe
```

**En Linux/macOS:**
```bash
./build.sh
./ana
```

---

## ¿Por qué no funciona `go run`?

Ana Streamer requiere compilación con CGO y el build tag `portaudio` para acceder a la librería de audio PortAudio. Los scripts de build automáticamente configuran esto:

```bash
go build -tags portaudio -o ana ./cmd/ana/main.go
```

`go run` no soporta los build tags ni la configuración CGO automáticamente.

---

## Compilación Manual (si lo necesitas)

**Windows (desde MSYS2 MinGW shell):**
```bash
set CGO_ENABLED=1
set CC=gcc
go build -tags portaudio -o ana.exe ./cmd/ana/main.go
```

**Linux/macOS:**
```bash
export CGO_ENABLED=1
go build -tags portaudio -o ana ./cmd/ana/main.go
```

---

## Verificar que funciona

Deberías ver:
```
═══════════════════════════════════════════════════════
🎤 Ana Streamer Active
═══════════════════════════════════════════════════════

🔊 How to use:
   1. Say 'Ana' to activate persistent session
   2. Keep talking - no need to repeat 'Ana'
   3. Say 'Adiós Ana' or similar to deactivate

💬 Deactivation words: adiós, detente, silencio, para ana,
   cállate, quieta, deja de grabar, stop, adiós ana

⌨️  Hotkey: F4 (press and hold to record)

Press Ctrl+C to exit
═══════════════════════════════════════════════════════
```

En este punto, Ana está escuchando:
- **Por voz**: Dí "Ana" para activar
- **Por hotkey**: Presiona F4 para grabar un comando

---

## Si aún falla con "PortAudio not found"

Asegúrate que PortAudio esté instalado:

**Windows (MSYS2):**
```bash
pacman -S mingw-w64-x86_64-portaudio
```

**Ubuntu/Debian:**
```bash
sudo apt-get install portaudio19-dev
```

**Fedora/RHEL:**
```bash
sudo dnf install portaudio-devel
```

**macOS:**
```bash
brew install portaudio
```

Luego ejecuta el script de build nuevamente.
