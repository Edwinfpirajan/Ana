# 🚀 Cómo Ejecutar Ana Streamer

## ⚠️ IMPORTANTE: Usa SIEMPRE los scripts de build

Ana Streamer requiere compilación especial con CGO y build tags. Los scripts de build manejan esto automáticamente.

### ✅ ESTO SÍ FUNCIONA - Método Recomendado

**En Windows:**
```bash
./build.bat
./ana.exe
```

**En Linux/macOS:**
```bash
./build.sh
./ana
```

Los scripts verifican automáticamente:
- ✅ Go está instalado
- ✅ PortAudio está disponible
- ✅ CGO está configurado
- ✅ Compilación con flags correctos
- ✅ Genera ejecutable optimizado

### ❌ ESTO NO FUNCIONA

```bash
# ❌ Error: "PortAudio build tag is required"
go run cmd/ana/main.go

# ❌ Error: "build constraints exclude all Go files"
go build -tags portaudio -o ana.exe ./cmd/ana/main.go  # (desde cmd/PowerShell normal)

# ❌ Error: falta CGO
go build -o ana ./cmd/ana/main.go
```

---

## ¿Por qué necesitan scripts especiales?

Ana Streamer necesita:
1. **CGO habilitado** - Para integración con código C
2. **Compilador GCC** - Para compilar PortAudio
3. **Build tag `portaudio`** - Para incluir audio capture
4. **Entorno MSYS2** (Windows) - Para que CGO funcione

Los scripts de build automáticamente configuran todo esto. Si lo haces manualmente, debes estar en MSYS2 MinGW shell:

**Windows (MSYS2 MinGW64 shell - ⚠️ REQUERIDO):**
```bash
# ⚠️ DEBE ejecutarse desde MSYS2 MinGW 64-bit shell
set CGO_ENABLED=1
set CC=gcc
go build -tags portaudio -o ana.exe ./cmd/ana/main.go
```

**Linux/macOS:**
```bash
export CGO_ENABLED=1
go build -tags portaudio -o ana ./cmd/ana/main.go
```

Pero es **MUCHO más fácil y más confiable** usar los scripts de build proporcionados.

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
