# ═══════════════════════════════════════════════════════════════════════════════
#                    DESCARGADOR DE VOCES PIPER (Windows)
# ═══════════════════════════════════════════════════════════════════════════════
# Script para descargar modelos de voz en español para Piper TTS

param(
    [string]$Language = "es_ES",  # es_ES (España) o es_MX (México)
    [string]$Voice = "davefx",    # davefx, mls, etc.
    [string]$Quality = "medium"   # low, medium, high
)

$ErrorActionPreference = "Stop"

Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  DESCARGADOR DE VOCES PIPER TTS" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Detectar directorio raíz del proyecto
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$VoicesDir = Join-Path $ProjectRoot "assets\voices\piper"

Write-Host "[1/5] Preparando directorios..." -ForegroundColor Yellow
if (-not (Test-Path $VoicesDir)) {
    New-Item -ItemType Directory -Path $VoicesDir -Force | Out-Null
}
Write-Host "  ✓ Directorio: $VoicesDir" -ForegroundColor Green

# Construir nombre del modelo
$ModelName = "$Language-$Voice-$Quality"
$BaseUrl = "https://huggingface.co/rhasspy/piper-voices/resolve/main"

# Convertir lenguaje a path (es_ES -> es/es_ES)
$LangCode = $Language.Split("_")[0]
$VoicePath = "$LangCode/$Language/$Voice/$Quality"

# Archivos a descargar
$OnnxFile = "$ModelName.onnx"
$JsonFile = "$ModelName.onnx.json"

$OnnxUrl = "$BaseUrl/$VoicePath/$OnnxFile"
$JsonUrl = "$BaseUrl/$VoicePath/$JsonFile"

$OnnxPath = Join-Path $VoicesDir $OnnxFile
$JsonPath = Join-Path $VoicesDir $JsonFile

Write-Host "[2/5] Información del modelo:" -ForegroundColor Yellow
Write-Host "  Idioma: $Language" -ForegroundColor Cyan
Write-Host "  Voz: $Voice" -ForegroundColor Cyan
Write-Host "  Calidad: $Quality" -ForegroundColor Cyan
Write-Host "  Nombre: $ModelName" -ForegroundColor Cyan
Write-Host ""

# Verificar si ya existe
if ((Test-Path $OnnxPath) -and (Test-Path $JsonPath)) {
    Write-Host "  ⚠️  El modelo ya existe. ¿Descargar de nuevo? (S/N)" -ForegroundColor Yellow
    $Response = Read-Host
    if ($Response -ne "S" -and $Response -ne "s") {
        Write-Host "  ℹ️  Descarga cancelada" -ForegroundColor Gray
        exit 0
    }
}

Write-Host "[3/5] Descargando modelo ONNX..." -ForegroundColor Yellow
Write-Host "  URL: $OnnxUrl" -ForegroundColor Gray
try {
    # Mostrar progreso de descarga
    $ProgressPreference = 'Continue'
    Invoke-WebRequest -Uri $OnnxUrl -OutFile $OnnxPath -UseBasicParsing

    $SizeMB = [math]::Round((Get-Item $OnnxPath).Length / 1MB, 2)
    Write-Host "  ✓ Descargado: $SizeMB MB" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Error al descargar modelo: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Posibles causas:" -ForegroundColor Yellow
    Write-Host "  - Combinación idioma/voz/calidad no existe" -ForegroundColor Gray
    Write-Host "  - Problema de conexión a internet" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Voces disponibles en español:" -ForegroundColor Cyan
    Write-Host "  • es_ES-davefx-medium (España, masculina) ← Recomendada" -ForegroundColor White
    Write-Host "  • es_ES-mls-medium (España, múltiples)" -ForegroundColor White
    Write-Host "  • es_MX-ald-medium (México, masculina)" -ForegroundColor White
    Write-Host ""
    exit 1
}

Write-Host "[4/5] Descargando configuración JSON..." -ForegroundColor Yellow
Write-Host "  URL: $JsonUrl" -ForegroundColor Gray
try {
    Invoke-WebRequest -Uri $JsonUrl -OutFile $JsonPath -UseBasicParsing
    Write-Host "  ✓ Descarga completada" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Error al descargar configuración: $_" -ForegroundColor Red
    exit 1
}

Write-Host "[5/5] Verificando integridad..." -ForegroundColor Yellow
if ((Test-Path $OnnxPath) -and (Test-Path $JsonPath)) {
    Write-Host "  ✓ Archivos verificados" -ForegroundColor Green
} else {
    Write-Host "  ✗ Error: Archivos incompletos" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "  ✓ VOZ DESCARGADA EXITOSAMENTE" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""
Write-Host "Archivos instalados:" -ForegroundColor Cyan
Write-Host "  • $OnnxFile" -ForegroundColor White
Write-Host "  • $JsonFile" -ForegroundColor White
Write-Host ""
Write-Host "Ubicación: $VoicesDir" -ForegroundColor Gray
Write-Host ""

# Mostrar configuración sugerida
Write-Host "Configuración para ana.config.yaml:" -ForegroundColor Yellow
Write-Host ""
Write-Host "tts:" -ForegroundColor White
Write-Host "  provider: `"piper`"" -ForegroundColor White
Write-Host "  piper:" -ForegroundColor White
Write-Host "    binary_path: `"./bin/piper/piper.exe`"" -ForegroundColor White
Write-Host "    model_path: `"./assets/voices/piper/$OnnxFile`"" -ForegroundColor Cyan
Write-Host "    speed: 1.0" -ForegroundColor White
Write-Host ""

# Probar la voz si Piper está instalado
$PiperExe = Join-Path $ProjectRoot "bin\piper\piper.exe"
if (Test-Path $PiperExe) {
    Write-Host "¿Quieres probar la voz ahora? (S/N)" -ForegroundColor Yellow
    $TestResponse = Read-Host
    if ($TestResponse -eq "S" -or $TestResponse -eq "s") {
        Write-Host ""
        Write-Host "Generando audio de prueba..." -ForegroundColor Yellow

        $TestText = "Hola, soy Ana, tu asistente de voz personal. Este es un test de la síntesis de voz en español."
        $OutputWav = Join-Path $ProjectRoot "test_voice.wav"

        try {
            # Ejecutar Piper para generar audio
            $TestText | & $PiperExe --model $OnnxPath --output_file $OutputWav

            if (Test-Path $OutputWav) {
                Write-Host "  ✓ Audio generado: $OutputWav" -ForegroundColor Green
                Write-Host ""
                Write-Host "Reproduciendo..." -ForegroundColor Cyan

                # Reproducir con PowerShell
                $player = New-Object System.Media.SoundPlayer $OutputWav
                $player.PlaySync()

                Write-Host "  ✓ Reproducción completada" -ForegroundColor Green

                # Limpiar archivo temporal
                Start-Sleep -Seconds 1
                Remove-Item $OutputWav -Force
            }
        } catch {
            Write-Host "  ⚠️  No se pudo reproducir el audio: $_" -ForegroundColor Yellow
            Write-Host "  Pero el modelo está instalado correctamente" -ForegroundColor Gray
        }
    }
}

Write-Host ""
Write-Host "Para descargar más voces, ejecuta:" -ForegroundColor Yellow
Write-Host "  .\scripts\download_voices.ps1 -Language es_MX -Voice ald -Quality medium" -ForegroundColor White
Write-Host ""
