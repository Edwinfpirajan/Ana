# ═══════════════════════════════════════════════════════════════════════════════
#                    VERIFICADOR DE DEPENDENCIAS - ANA
# ═══════════════════════════════════════════════════════════════════════════════
# Script para verificar que todos los componentes estén instalados y configurados

$ErrorActionPreference = "Continue"

function Write-Status {
    param(
        [string]$Component,
        [bool]$IsOk,
        [string]$Message = ""
    )

    $Status = if ($IsOk) { "✅" } else { "❌" }
    $Color = if ($IsOk) { "Green" } else { "Red" }

    Write-Host "  [$Status] " -NoNewline -ForegroundColor $Color
    Write-Host "$Component" -NoNewline -ForegroundColor White

    if ($Message) {
        Write-Host " - $Message" -ForegroundColor Gray
    } else {
        Write-Host ""
    }
}

function Write-Info {
    param([string]$Text)
    Write-Host "      ℹ️  $Text" -ForegroundColor Cyan
}

Clear-Host
Write-Host @"
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║         🔍 VERIFICADOR DE DEPENDENCIAS - ANA 🔍            ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
"@ -ForegroundColor Cyan

$ProjectRoot = Split-Path -Parent $PSScriptRoot

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "  REQUISITOS DEL SISTEMA" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

# Go
$GoInstalled = $false
try {
    $GoVersion = go version 2>&1
    if ($LASTEXITCODE -eq 0) {
        $GoInstalled = $true
        Write-Status "Go" $true $GoVersion
    } else {
        Write-Status "Go" $false "No instalado"
        Write-Info "Instala desde: https://go.dev/dl/"
    }
} catch {
    Write-Status "Go" $false "No instalado"
    Write-Info "Instala desde: https://go.dev/dl/"
}

# PowerShell
$PSVersion = $PSVersionTable.PSVersion
$PSVersionOk = $PSVersion.Major -ge 5
Write-Status "PowerShell" $PSVersionOk "v$($PSVersion.Major).$($PSVersion.Minor)"

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "  COMPONENTES LOCALES (STT, LLM, TTS)" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

# Ollama (LLM)
Write-Host "▶ Ollama (LLM)" -ForegroundColor Cyan
$OllamaInstalled = $false
$OllamaRunning = $false
$OllamaModel = $false

try {
    $OllamaVer = ollama --version 2>&1
    if ($LASTEXITCODE -eq 0 -or $OllamaVer -match "ollama version") {
        $OllamaInstalled = $true
        Write-Status "  Binario" $true $OllamaVer

        # Verificar si está corriendo
        try {
            $Response = Invoke-WebRequest -Uri "http://localhost:11434/api/version" -UseBasicParsing -TimeoutSec 2 -ErrorAction SilentlyContinue
            $OllamaRunning = $true
            Write-Status "  Servicio" $true "Corriendo en puerto 11434"

            # Verificar modelo
            try {
                $Models = ollama list 2>&1 | Out-String
                if ($Models -match "llama3\.2.*3b") {
                    $OllamaModel = $true
                    Write-Status "  Modelo llama3.2:3b" $true "Instalado"
                } else {
                    Write-Status "  Modelo llama3.2:3b" $false "No encontrado"
                    Write-Info "Ejecuta: ollama pull llama3.2:3b"
                }

                # Mostrar otros modelos
                $ModelLines = $Models -split "`n" | Select-Object -Skip 1 | Where-Object { $_.Trim() -ne "" }
                if ($ModelLines.Count -gt 0) {
                    Write-Info "Modelos instalados: $($ModelLines.Count)"
                }
            } catch {
                Write-Status "  Modelo" $false "No se pudo verificar"
            }
        } catch {
            $OllamaRunning = $false
            Write-Status "  Servicio" $false "No está corriendo"
            Write-Info "Ejecuta en otra terminal: ollama serve"
        }
    } else {
        Write-Status "  Binario" $false "No instalado"
        Write-Info "Instala desde: https://ollama.ai/download"
    }
} catch {
    Write-Status "  Binario" $false "No instalado"
    Write-Info "Instala desde: https://ollama.ai/download"
}

Write-Host ""

# Piper (TTS)
Write-Host "▶ Piper (TTS)" -ForegroundColor Cyan
$PiperExe = Join-Path $ProjectRoot "bin\piper\piper.exe"
$PiperInstalled = Test-Path $PiperExe

if ($PiperInstalled) {
    Write-Status "  Binario" $true $PiperExe

    # Verificar versión
    try {
        $PiperVer = & $PiperExe --version 2>&1
        Write-Info "Versión: $PiperVer"
    } catch {
        Write-Info "No se pudo verificar versión"
    }
} else {
    Write-Status "  Binario" $false "No encontrado"
    Write-Info "Ejecuta: .\scripts\install_piper.ps1"
}

# Verificar voces
$VoicesDir = Join-Path $ProjectRoot "assets\voices\piper"
$Voices = Get-ChildItem -Path $VoicesDir -Filter "*.onnx" -ErrorAction SilentlyContinue

if ($Voices -and $Voices.Count -gt 0) {
    Write-Status "  Voces" $true "$($Voices.Count) voz(es) instalada(s)"
    foreach ($Voice in $Voices) {
        $SizeMB = [math]::Round($Voice.Length / 1MB, 1)
        Write-Info "$($Voice.Name) ($SizeMB MB)"
    }
} else {
    Write-Status "  Voces" $false "Sin voces instaladas"
    Write-Info "Ejecuta: .\scripts\download_voices.ps1"
}

Write-Host ""

# Whisper (STT)
Write-Host "▶ Whisper.cpp (STT)" -ForegroundColor Cyan
$WhisperExe = Join-Path $ProjectRoot "bin\whisper\main.exe"
$WhisperInstalled = Test-Path $WhisperExe

if ($WhisperInstalled) {
    Write-Status "  Binario" $true $WhisperExe

    # Verificar versión
    try {
        $WhisperHelp = & $WhisperExe --help 2>&1
        if ($WhisperHelp -match "usage:") {
            Write-Info "Ejecutable funcional"
        }
    } catch {
        Write-Status "  Binario" $false "No ejecutable"
    }
} else {
    Write-Status "  Binario" $false "No encontrado"
    Write-Info "Descarga desde: https://github.com/ggerganov/whisper.cpp/releases"
    Write-Info "Extrae main.exe en: bin\whisper\"
}

# Verificar modelos
$ModelsDir = Join-Path $ProjectRoot "assets\models\whisper"
$Models = Get-ChildItem -Path $ModelsDir -Filter "ggml-*.bin" -ErrorAction SilentlyContinue

if ($Models -and $Models.Count -gt 0) {
    Write-Status "  Modelos" $true "$($Models.Count) modelo(s) instalado(s)"
    foreach ($Model in $Models) {
        $SizeMB = [math]::Round($Model.Length / 1MB, 1)
        Write-Info "$($Model.Name) ($SizeMB MB)"
    }
} else {
    Write-Status "  Modelos" $false "Sin modelos descargados"
    Write-Info "Ejecuta: .\scripts\install_whisper.ps1"
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "  CONFIGURACIÓN" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

# Verificar archivo de configuración
$ConfigFile = Join-Path $ProjectRoot "config\ana.config.yaml"
$ConfigExists = Test-Path $ConfigFile

Write-Status "Archivo de config" $ConfigExists $ConfigFile

if ($ConfigExists) {
    $ConfigContent = Get-Content $ConfigFile -Raw

    # Parsear proveedores
    if ($ConfigContent -match 'stt:\s+provider:\s*"(\w+)"') {
        $SttProvider = $matches[1]
        $SttLocal = $SttProvider -eq "whisper"
        Write-Status "  STT Provider" $SttLocal "$SttProvider $(if($SttLocal){'(local)'}else{'(cloud)'})"
    }

    if ($ConfigContent -match 'llm:\s+provider:\s*"(\w+)"') {
        $LlmProvider = $matches[1]
        $LlmLocal = $LlmProvider -eq "ollama"
        Write-Status "  LLM Provider" $LlmLocal "$LlmProvider $(if($LlmLocal){'(local)'}else{'(cloud)'})"
    }

    if ($ConfigContent -match 'tts:\s+provider:\s*"(\w+)"') {
        $TtsProvider = $matches[1]
        $TtsLocal = $TtsProvider -eq "piper"
        Write-Status "  TTS Provider" $TtsLocal "$TtsProvider $(if($TtsLocal){'(local)'}else{'(cloud)'})"
    }

    # Verificar rutas de binarios
    Write-Host ""
    Write-Host "  Rutas configuradas:" -ForegroundColor Cyan

    if ($ConfigContent -match 'binary_path:\s*"([^"]+)"') {
        $PiperPath = $matches[1]
        Write-Host "    Piper: $PiperPath" -ForegroundColor Gray
    }

    if ($ConfigContent -match 'whisper:.*?binary_path:\s*"([^"]+)"') {
        $WhisperPath = $matches[1]
        Write-Host "    Whisper: $WhisperPath" -ForegroundColor Gray
    }
} else {
    Write-Info "Crea desde: config\ana.config.example.yaml"
}

# Verificar .env
$EnvFile = Join-Path $ProjectRoot ".env"
$EnvExists = Test-Path $EnvFile
Write-Status ".env file" $EnvExists $EnvFile

if ($EnvExists) {
    $EnvContent = Get-Content $EnvFile -Raw
    $HasApiKey = $EnvContent -match "OPENAI_API_KEY"
    if ($HasApiKey) {
        Write-Info "OPENAI_API_KEY configurada (para fallback)"
    }
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "  COMPILACIÓN" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

# Verificar go.mod
$GoMod = Join-Path $ProjectRoot "go.mod"
$GoModExists = Test-Path $GoMod
Write-Status "go.mod" $GoModExists

# Verificar main
$MainGo = Join-Path $ProjectRoot "cmd\ana\main.go"
$MainExists = Test-Path $MainGo
Write-Status "main.go" $MainExists $MainGo

# Verificar si ya está compilado
$AnaExe = Join-Path $ProjectRoot "ana.exe"
$IsCompiled = Test-Path $AnaExe

if ($IsCompiled) {
    $FileInfo = Get-Item $AnaExe
    $SizeMB = [math]::Round($FileInfo.Length / 1MB, 2)
    $Modified = $FileInfo.LastWriteTime.ToString("yyyy-MM-dd HH:mm")
    Write-Status "ana.exe" $true "$SizeMB MB (compilado: $Modified)"
} else {
    Write-Status "ana.exe" $false "No compilado"
    Write-Info "Ejecuta: go build -o ana.exe ./cmd/ana"
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "  RESUMEN" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""

# Calcular puntuación
$Score = 0
$Total = 0

# Componentes críticos
$Total += 1; if ($OllamaInstalled) { $Score += 1 }
$Total += 1; if ($OllamaRunning) { $Score += 1 }
$Total += 1; if ($OllamaModel) { $Score += 1 }
$Total += 1; if ($PiperInstalled) { $Score += 1 }
$Total += 1; if ($Voices -and $Voices.Count -gt 0) { $Score += 1 }
$Total += 1; if ($WhisperInstalled) { $Score += 1 }
$Total += 1; if ($Models -and $Models.Count -gt 0) { $Score += 1 }
$Total += 1; if ($ConfigExists) { $Score += 1 }

$Percentage = [math]::Round(($Score / $Total) * 100)

Write-Host "Componentes listos: $Score / $Total ($Percentage%)" -ForegroundColor Cyan
Write-Host ""

if ($Percentage -eq 100) {
    Write-Host "🎉 ¡TODO LISTO! Ana está configurado al 100%" -ForegroundColor Green
    Write-Host ""
    Write-Host "Próximo paso:" -ForegroundColor Yellow
    Write-Host "  1. Compilar: go build -o ana.exe ./cmd/ana" -ForegroundColor White
    Write-Host "  2. Ejecutar: .\ana.exe" -ForegroundColor White
} elseif ($Percentage -ge 75) {
    Write-Host "👍 ¡Casi listo! Solo faltan algunos componentes" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Próximo paso:" -ForegroundColor Yellow
    Write-Host "  Ejecuta: .\scripts\setup_local.ps1" -ForegroundColor White
} elseif ($Percentage -ge 50) {
    Write-Host "⚠️  Estás a mitad de camino" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Próximo paso:" -ForegroundColor Yellow
    Write-Host "  Ejecuta: .\scripts\setup_local.ps1" -ForegroundColor White
} else {
    Write-Host "❌ Muchos componentes faltantes" -ForegroundColor Red
    Write-Host ""
    Write-Host "Próximo paso:" -ForegroundColor Yellow
    Write-Host "  1. Ejecuta: .\scripts\setup_local.ps1" -ForegroundColor White
    Write-Host "  2. Consulta: SETUP_LOCAL.md" -ForegroundColor White
}

Write-Host ""
Write-Host "Para más información, consulta:" -ForegroundColor Cyan
Write-Host "  • QUICKSTART.md  - Inicio rápido" -ForegroundColor White
Write-Host "  • SETUP_LOCAL.md - Guía detallada" -ForegroundColor White
Write-Host "  • scripts\README.md - Referencia de scripts" -ForegroundColor White
Write-Host ""
