# 🚀 Quickstart - Ana Local Setup

## ⚡ Setup en 5 Minutos

### 1. Ejecutar Script Automático

```powershell
.\scripts\setup_local.ps1
```

### 2. Instalar Ollama

```powershell
# Descargar desde: https://ollama.ai/download
# O con Winget:
winget install Ollama.Ollama

# Iniciar Ollama
ollama serve

# En otra terminal, descargar modelo
ollama pull llama3.2:3b
```

### 3. Compilar y Ejecutar

```powershell
# Compilar
go build -o ana.exe ./cmd/ana

# Ejecutar
.\ana.exe
```

---

## ✅ Verificación Rápida

```powershell
# Ollama corriendo?
ollama list

# Piper instalado?
.\bin\piper\piper.exe --version

# Configuración correcta?
cat config\ana.config.yaml | Select-String "provider"
```

---  

## 🌐 ¿Piper falla?

- Si `Ana` sigue sin hablar y el log dice `piper execution failed: exit status 0xc0000409`, cambia temporalmente el bloque `tts` en `config/ana.config.yaml` a `provider: "openai"`.  
- Asegúrate de cargar tu `OPENAI_API_KEY` desde `.env` (usa `.\load_env.ps1`) antes de ejecutar el comando; así el fallback cloud toma la voz automáticamente.  
- Cuando tengas un build de Piper que no crashée, vuelve a `provider: "auto"` o `provider: "piper"` para priorizar el TTS local.

---

## 📖 Documentación Completa

Para instalación detallada, consulta: [SETUP_LOCAL.md](SETUP_LOCAL.md)

---

## 🎯 Configuración Actual

Tu Ana está configurado para:

- **STT**: Whisper.cpp (local)
- **LLM**: Ollama (local)
- **TTS**: Piper (local)

**Modo 100% local - Sin dependencias de IA en la nube** ✅

---

## ⚠️ Nota Importante

**Whisper.cpp** requiere descargar el binario precompilado manualmente:

1. Visita: https://github.com/ggerganov/whisper.cpp/releases
2. Descarga: `whisper-bin-x64.zip`
3. Extrae `main.exe` en: `bin\whisper\`

**Alternativa temporal**: Usa OpenAI STT cambiando en config:

```yaml
stt:
  provider: "openai"
```

---

## 🆘 Problemas Comunes

| Problema | Solución |
|----------|----------|
| "Ollama not running" | `ollama serve` en otra terminal |
| "Piper not found" | `.\scripts\install_piper.ps1` |
| "Model not found" | Verifica rutas en `ana.config.yaml` |

---

**¡Listo para hablar con Ana!** 🎙️
