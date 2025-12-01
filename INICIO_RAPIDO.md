# 🚀 Inicio Rápido - Ana Streamer

## Opción 1: Usar el Repo Existente (Local)

```bash
cd c:\Users\Ferchando\Documents\ana
.\ana.exe
```

Luego di: **"Ana, ¿qué hora es?"**

## Opción 2: Clonar Repo Limpio desde GitHub

```bash
git clone -b ana https://github.com/Edwinfpirajan/Ana.git
cd Ana
```

### Descargar Modelos (si aún no los tienes)

```bash
# Opción A: Desde HuggingFace
# Whisper (141 MB): https://huggingface.co/ggerganov/whisper.cpp
# Piper voices (60-108 MB cada): https://huggingface.co/rhasspy/piper

# Opción B: Desde rama main del repo
git checkout main
# Copiar assets/
git checkout ana
```

### Compilar

```bash
# Opción 1: Script
.\build.bat

# Opción 2: Manual con bash
bash -c "export CGO_ENABLED=1 CC=/c/msys64/mingw64/bin/gcc && \
         go build -tags portaudio -o ana.exe ./cmd/ana/main.go"
```

### Ejecutar

```bash
.\ana.exe
```

## Requisitos Previos

- ✅ Go 1.20+ (instalado)
- ✅ MSYS2 + GCC (ya configurado)
- ✅ PortAudio headers (en MSYS2)
- ✅ Ollama (para LLM local - opcional pero recomendado)

## Comandos Útiles

```bash
# Verificar componentes
.\test_components.exe

# Ver logs detallados
set LOG_LEVEL=debug
.\ana.exe

# Compilar y ejecutar inmediatamente
.\build.bat && .\ana.exe
```

## ¿Problemas?

1. Verifica que Ollama está corriendo:
   ```bash
   ollama serve
   ```

2. Verifica componentes:
   ```bash
   .\test_components.exe
   ```

3. Lee SETUP_ANA.md para instrucciones detalladas

## Contacto

GitHub: https://github.com/Edwinfpirajan/Ana
Rama principal: ana
