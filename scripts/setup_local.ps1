# ═══════════════════════════════════════════════════════════════════════════════
#                    SETUP COMPLETO - ANA 100% LOCAL
# ═══════════════════════════════════════════════════════════════════════════════
# Script maestro para configurar Ana completamente local sin dependencias de IA

param(
    [switch]$SkipPiper,
    [switch]$SkipWhisper,
    [switch]$SkipOllama,
    [string]$WhisperModel = "base",
    [string]$PiperVoice = "es_ES-davefx-medium"
)

$ErrorActionPreference = "Stop"

# Colores
function Write-Title {
    param([string]$Text)
    Write-Host ""
    Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  $Text" -ForegroundColor Cyan
    Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host ""
}

function Write-Step {
    param([string]$Text)
    Write-Host "▶ $Text" -ForegroundColor Yellow
}

function Write-Success {
    param([string]$Text)
    Write-Host "  ✓ $Text" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Text)
    Write-Host "  ⚠️  $Text" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Text)
    Write-Host "  ✗ $Text" -ForegroundColor Red
}

function Write-Info {
    param([string]$Text)
    Write-Host "  ℹ️  $Text" -ForegroundColor Cyan
}

# Banner
Clear-Host
Write-Host @"
    ╔═══════════════════════════════════════════════════════════════╗
    ║                                                               ║
    ║          🎙️  ANA STREAMER - SETUP LOCAL 100%  🎙️           ║
    ║                                                               ║
    ║              Sin dependencias de IA en la nube                ║
    ║                                                               ║
    ╚═══════════════════════════════════════════════════════════════╝
"@ -ForegroundColor Cyan

Write-Host ""
Write-Host "Este script instalará y configurará:" -ForegroundColor White
Write-Host "  • Piper TTS (síntesis de voz local)" -ForegroundColor Gray
Write-Host "  • Whisper.cpp (reconocimiento de voz local)" -ForegroundColor Gray
Write-Host "  • Ollama (modelo de lenguaje local)" -ForegroundColor Gray
Write-Host "  • Configuración de Ana optimizada para uso local" -ForegroundColor Gray
Write-Host ""

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$ScriptsDir = Join-Path $ProjectRoot "scripts"

# ═════════════════════════════════════════════════════════════════════════════
# PASO 1: VERIFICAR OLLAMA
# ═════════════════════════════════════════════════════════════════════════════
Write-Title "PASO 1/4 - VERIFICAR OLLAMA"

if (-not $SkipOllama) {
    Write-Step "Verificando si Ollama está instalado..."

    $OllamaInstalled = $false
    try {
        $OllamaVersion = ollama --version 2>&1
        if ($LASTEXITCODE -eq 0 -or $OllamaVersion -match "ollama version") {
            $OllamaInstalled = $true
            Write-Success "Ollama instalado: $OllamaVersion"
        }
    } catch {
        $OllamaInstalled = $false
    }

    if (-not $OllamaInstalled) {
        Write-Warning "Ollama NO está instalado"
        Write-Info "Descarga Ollama desde: https://ollama.ai/download"
        Write-Host ""
        Write-Host "Pasos para instalar Ollama:" -ForegroundColor Yellow
        Write-Host "  1. Visita: https://ollama.ai/download" -ForegroundColor White
        Write-Host "  2. Descarga el instalador para Windows" -ForegroundColor White
        Write-Host "  3. Ejecuta el instalador" -ForegroundColor White
        Write-Host "  4. Abre PowerShell y ejecuta: ollama serve" -ForegroundColor White
        Write-Host "  5. En otra terminal ejecuta: ollama pull llama3.2:3b" -ForegroundColor White
        Write-Host ""
        Write-Host "¿Continuar sin Ollama? (S/N) [Puedes instalarlo después]" -ForegroundColor Cyan
        $Response = Read-Host
        if ($Response -ne "S" -and $Response -ne "s") {
            Write-Info "Setup cancelado. Instala Ollama y vuelve a ejecutar este script."
            exit 0
        }
    } else {
        Write-Step "Verificando si Ollama está corriendo..."
        try {
            $OllamaRunning = Invoke-WebRequest -Uri "http://localhost:11434/api/version" -UseBasicParsing -ErrorAction SilentlyContinue
            Write-Success "Ollama está corriendo"

            Write-Step "Verificando modelo llama3.2:3b..."
            try {
                $Models = ollama list 2>&1 | Out-String
                if ($Models -match "llama3.2.*3b") {
                    Write-Success "Modelo llama3.2:3b está instalado"
                } else {
                    Write-Warning "Modelo llama3.2:3b NO encontrado"
                    Write-Info "Descargando modelo... (esto puede tardar varios minutos)"
                    ollama pull llama3.2:3b
                    if ($LASTEXITCODE -eq 0) {
                        Write-Success "Modelo descargado exitosamente"
                    }
                }
            } catch {
                Write-Warning "No se pudo verificar modelos: $_"
            }
        } catch {
            Write-Warning "Ollama instalado pero NO está corriendo"
            Write-Info "Ejecuta en otra terminal: ollama serve"
        }
    }
} else {
    Write-Info "Verificación de Ollama omitida (--SkipOllama)"
}

# ═════════════════════════════════════════════════════════════════════════════
# PASO 2: INSTALAR PIPER TTS
# ═════════════════════════════════════════════════════════════════════════════
Write-Title "PASO 2/4 - INSTALAR PIPER TTS"

if (-not $SkipPiper) {
    Write-Step "Ejecutando instalador de Piper..."

    $PiperScript = Join-Path $ScriptsDir "install_piper.ps1"
    if (Test-Path $PiperScript) {
        & $PiperScript
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Piper instalado correctamente"

            # Descargar voz
            Write-Step "Descargando voz en español..."
            $VoiceScript = Join-Path $ScriptsDir "download_voices.ps1"

            # Parsear nombre de voz
            $VoiceParts = $PiperVoice.Split("-")
            if ($VoiceParts.Length -ge 3) {
                $Lang = $VoiceParts[0]
                $Voice = $VoiceParts[1]
                $Quality = $VoiceParts[2]

                & $VoiceScript -Language $Lang -Voice $Voice -Quality $Quality
                if ($LASTEXITCODE -eq 0) {
                    Write-Success "Voz descargada correctamente"
                }
            }
        }
    } else {
        Write-Error-Custom "No se encontró install_piper.ps1"
    }
} else {
    Write-Info "Instalación de Piper omitida (--SkipPiper)"
}

# ═════════════════════════════════════════════════════════════════════════════
# PASO 3: INSTALAR WHISPER.CPP (MODELOS)
# ═════════════════════════════════════════════════════════════════════════════
Write-Title "PASO 3/4 - CONFIGURAR WHISPER.CPP"

if (-not $SkipWhisper) {
    Write-Step "Ejecutando instalador de Whisper..."

    $WhisperScript = Join-Path $ScriptsDir "install_whisper.ps1"
    if (Test-Path $WhisperScript) {
        & $WhisperScript -Model $WhisperModel
        Write-Info "Whisper.cpp requiere binario compilado o precompilado"
        Write-Info "Consulta: https://github.com/ggerganov/whisper.cpp/releases"
    } else {
        Write-Error-Custom "No se encontró install_whisper.ps1"
    }
} else {
    Write-Info "Instalación de Whisper omitida (--SkipWhisper)"
}

# ═════════════════════════════════════════════════════════════════════════════
# PASO 4: VERIFICAR CONFIGURACIÓN
# ═════════════════════════════════════════════════════════════════════════════
Write-Title "PASO 4/4 - VERIFICAR CONFIGURACIÓN"

$ConfigFile = Join-Path $ProjectRoot "config\ana.config.yaml"
if (Test-Path $ConfigFile) {
    Write-Success "Configuración encontrada: $ConfigFile"

    Write-Step "Verificando proveedores configurados..."
    $ConfigContent = Get-Content $ConfigFile -Raw

    # Verificar STT
    if ($ConfigContent -match 'stt:\s+provider:\s*"(\w+)"') {
        $SttProvider = $matches[1]
        Write-Info "STT Provider: $SttProvider"
        if ($SttProvider -eq "whisper") {
            Write-Success "STT configurado para modo local"
        } elseif ($SttProvider -eq "openai") {
            Write-Warning "STT usa OpenAI (requiere API key y conexión)"
        }
    }

    # Verificar LLM
    if ($ConfigContent -match 'llm:\s+provider:\s*"(\w+)"') {
        $LlmProvider = $matches[1]
        Write-Info "LLM Provider: $LlmProvider"
        if ($LlmProvider -eq "ollama") {
            Write-Success "LLM configurado para modo local (Ollama)"
        } elseif ($LlmProvider -eq "openai") {
            Write-Warning "LLM usa OpenAI (requiere API key y conexión)"
        }
    }

    # Verificar TTS
    if ($ConfigContent -match 'tts:\s+provider:\s*"(\w+)"') {
        $TtsProvider = $matches[1]
        Write-Info "TTS Provider: $TtsProvider"
        if ($TtsProvider -eq "piper") {
            Write-Success "TTS configurado para modo local (Piper)"
        } elseif ($TtsProvider -eq "openai") {
            Write-Warning "TTS usa OpenAI (requiere API key y conexión)"
        }
    }
} else {
    Write-Warning "No se encontró ana.config.yaml"
    Write-Info "Crea uno desde: config/ana.config.example.yaml"
}

# ═════════════════════════════════════════════════════════════════════════════
# RESUMEN FINAL
# ═════════════════════════════════════════════════════════════════════════════
Write-Title "RESUMEN DE INSTALACIÓN"

Write-Host "Estado de componentes:" -ForegroundColor Cyan
Write-Host ""

# Verificar Ollama
$OllamaStatus = "❌ No instalado"
try {
    $null = ollama --version 2>&1
    if ($LASTEXITCODE -eq 0) {
        $OllamaStatus = "✅ Instalado"
    }
} catch {}
Write-Host "  Ollama (LLM):       $OllamaStatus" -ForegroundColor White

# Verificar Piper
$PiperExe = Join-Path $ProjectRoot "bin\piper\piper.exe"
$PiperStatus = if (Test-Path $PiperExe) { "✅ Instalado" } else { "❌ No instalado" }
Write-Host "  Piper (TTS):        $PiperStatus" -ForegroundColor White

# Verificar voces Piper
$VoicesDir = Join-Path $ProjectRoot "assets\voices\piper"
$Voices = Get-ChildItem -Path $VoicesDir -Filter "*.onnx" -ErrorAction SilentlyContinue
$VoicesStatus = if ($Voices) { "✅ $($Voices.Count) voz(es)" } else { "❌ Sin voces" }
Write-Host "  Voces Piper:        $VoicesStatus" -ForegroundColor White

# Verificar Whisper
$WhisperExe = Join-Path $ProjectRoot "bin\whisper\main.exe"
$WhisperStatus = if (Test-Path $WhisperExe) { "✅ Instalado" } else { "❌ No instalado" }
Write-Host "  Whisper (STT):      $WhisperStatus" -ForegroundColor White

# Verificar modelos Whisper
$ModelsDir = Join-Path $ProjectRoot "assets\models\whisper"
$Models = Get-ChildItem -Path $ModelsDir -Filter "ggml-*.bin" -ErrorAction SilentlyContinue
$ModelsStatus = if ($Models) { "✅ $($Models.Count) modelo(s)" } else { "❌ Sin modelos" }
Write-Host "  Modelos Whisper:    $ModelsStatus" -ForegroundColor White

Write-Host ""
Write-Host "Configuración actual:" -ForegroundColor Cyan
Write-Host "  • STT: $SttProvider" -ForegroundColor White
Write-Host "  • LLM: $LlmProvider" -ForegroundColor White
Write-Host "  • TTS: $TtsProvider" -ForegroundColor White

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "  SETUP COMPLETADO" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""

# Próximos pasos
Write-Host "Próximos pasos:" -ForegroundColor Yellow
Write-Host ""

if ($OllamaStatus -match "No instalado") {
    Write-Host "  1. Instalar Ollama desde: https://ollama.ai/download" -ForegroundColor White
}

if (-not (Test-Path $WhisperExe)) {
    Write-Host "  2. Descargar Whisper.cpp binario desde:" -ForegroundColor White
    Write-Host "     https://github.com/ggerganov/whisper.cpp/releases" -ForegroundColor Gray
}

if ($OllamaStatus -match "Instalado") {
    Write-Host "  3. Asegúrate de que Ollama esté corriendo:" -ForegroundColor White
    Write-Host "     ollama serve" -ForegroundColor Gray
}

Write-Host "  4. Compilar Ana:" -ForegroundColor White
Write-Host "     go build -o ana.exe ./cmd/ana" -ForegroundColor Gray

Write-Host "  5. Ejecutar Ana:" -ForegroundColor White
Write-Host "     .\ana.exe" -ForegroundColor Gray

Write-Host ""
Write-Host "Para más ayuda, consulta: README.md" -ForegroundColor Cyan
Write-Host ""
