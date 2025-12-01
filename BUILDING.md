# 🔨 Guía de Compilación - Ana Streamer

Esta guía explica cómo compilar Ana Streamer correctamente en diferentes sistemas operativos.

## ⚡ Compilación Rápida

### Windows
```bash
./build.bat
```

### Linux/macOS
```bash
./build.sh
```

---

## 📋 Requisitos

Ana Streamer requiere:
- **Go 1.22+**
- **GCC** (compilador C)
- **PortAudio** (librerías de desarrollo)
- **CGO habilitado** (para integración con C)

### Instalar Requisitos

#### Windows (MSYS2/MinGW)

```bash
# 1. Instalar PortAudio en MSYS2
pacman -S mingw-w64-x86_64-portaudio mingw-w64-x86_64-gcc

# 2. Verificar instalación
gcc --version
pkg-config --modversion portaudio-2.0
```

#### Ubuntu/Debian

```bash
# 1. Actualizar package manager
sudo apt-get update

# 2. Instalar dependencias
sudo apt-get install -y \
  build-essential \
  portaudio19-dev \
  gcc \
  pkg-config

# 3. Verificar instalación
gcc --version
pkg-config --modversion portaudio-2.0
```

#### Fedora/RHEL

```bash
# 1. Instalar dependencias
sudo dnf install -y \
  gcc \
  make \
  portaudio-devel \
  pkgconfig

# 2. Verificar instalación
gcc --version
pkg-config --modversion portaudio-2.0
```

#### macOS (Homebrew)

```bash
# 1. Instalar dependencias
brew install portaudio gcc

# 2. Verificar instalación
gcc --version
pkg-config --modversion portaudio-2.0
```

---

## 🔨 Compilar Manualmente

### Opción 1: Usando Scripts (Recomendado)

**Windows:**
```bash
./build.bat
```

**Linux/macOS:**
```bash
./build.sh
```

Los scripts:
- Verifican que Go esté instalado
- Validan las dependencias de PortAudio
- Configuran variables de entorno automáticamente
- Compilan con flags optimizados
- Muestran el tamaño del binario resultante

### Opción 2: Compilación Manual

#### Windows (MSYS2 Shell)

```bash
# Desde MSYS2 MinGW 64-bit shell:
cd /c/Users/Ferchando/Documents/ana
set CGO_ENABLED=1
set CC=gcc
go build -tags portaudio -o ana.exe ./cmd/ana/main.go
```

#### Linux/macOS

```bash
# Desde bash/zsh:
cd ~/Documents/ana
export CGO_ENABLED=1
go build -tags portaudio -o ana ./cmd/ana/main.go
```

---

## ✅ Verificar Compilación

Después de compilar, verifica que el ejecutable funciona:

### Windows
```bash
./ana.exe

# Deberías ver:
# [INFO] Ana Streamer starting...
# [INFO] Initializing Whisper STT provider
# [INFO] Initializing Ollama LLM provider
# ...
# 🎤 Ana Streamer Active
```

### Linux/macOS
```bash
./ana

# Deberías ver:
# [INFO] Ana Streamer starting...
# [INFO] Initializing Whisper STT provider
# [INFO] Initializing Ollama LLM provider
# ...
# 🎤 Ana Streamer Active
```

---

## 🛠️ Troubleshooting

### Error: "PortAudio build tag is required"

**Causa:** Compilaste sin el tag `-tags portaudio`

**Solución:**
```bash
# Asegúrate de usar:
go build -tags portaudio -o ana ./cmd/ana/main.go

# O usa el script de build:
./build.bat    # Windows
./build.sh     # Linux/macOS
```

### Error: "C compiler gcc not found"

**Causa:** GCC no está instalado o no está en PATH

**Solución:**
- **Windows:** Instala MinGW64 desde MSYS2: `pacman -S mingw-w64-x86_64-gcc`
- **Linux:** `sudo apt-get install build-essential`
- **macOS:** `brew install gcc`

### Error: "build constraints exclude all Go files in portaudio"

**Causa:** PortAudio Go binding no está disponible para tu sistema

**Solución 1:** Verifica CGO está habilitado
```bash
go env CGO_ENABLED  # Debe ser: 1
```

**Solución 2:** Instala PortAudio development files
```bash
# Ubuntu:
sudo apt-get install portaudio19-dev

# macOS:
brew install portaudio

# Windows/MSYS2:
pacman -S mingw-w64-x86_64-portaudio
```

### Error: "pkg-config not found"

**Causa:** pkg-config no está instalado

**Solución:**
- **Linux:** `sudo apt-get install pkg-config`
- **macOS:** `brew install pkg-config`
- **Windows/MSYS2:** `pacman -S pkg-config`

---

## 📦 Compilación Alternativa (sin PortAudio)

Si no necesitas captura de audio por ahora, puedes compilar sin PortAudio (mode stub):

```bash
# Esto compilará pero mostrará error "PortAudio is required":
go build -o ana ./cmd/ana/main.go
```

**Nota:** Esto es útil solo para testing. Para usar Ana Streamer, necesitas compilar con PortAudio.

---

## 🚀 Compilación Optimizada para Producción

Para crear un binario optimizado (más pequeño y rápido):

### Windows
```bash
set CGO_ENABLED=1
set CC=gcc
go build -tags portaudio -ldflags="-s -w" -o ana.exe ./cmd/ana/main.go
```

### Linux/macOS
```bash
export CGO_ENABLED=1
go build -tags portaudio -ldflags="-s -w" -o ana ./cmd/ana/main.go
```

**Flags explicados:**
- `-ldflags="-s -w"` remueve símbolos de debug (reduce ~20% de tamaño)
- Usa `UPX` para comprimir aún más (opcional)

---

## 🧹 Limpiar Builds Anteriores

```bash
# Eliminar binarios
rm -f ana ana.exe

# Limpiar cache de módulos
go clean -modcache

# Limpiar cache de compilación
go clean -cache

# Limpiar todo
go clean -modcache && rm -f ana ana.exe
```

---

## 🔗 Variables de Entorno Útiles

```bash
# Ver estado de CGO
go env CGO_ENABLED

# Ver compilador C configurado
go env CC

# Ver todos los flags de compilación
go env

# Forzar recompilación
go clean -a
```

---

## 📚 Referencias

- [Go CGO Documentation](https://golang.org/cmd/cgo/)
- [PortAudio Official Site](http://www.portaudio.com/)
- [portaudio Go Binding](https://github.com/gordonklaus/portaudio)

---

¿Problemas? Abre un issue en el repositorio con:
- Tu sistema operativo y versión
- Salida de `go version` y `gcc --version`
- Error completo de compilación
