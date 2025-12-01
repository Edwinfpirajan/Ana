# ═══════════════════════════════════════════════════════════════════════════════
#                    INSTALADOR DE WHISPER.CPP (Windows)
# ═══════════════════════════════════════════════════════════════════════════════
# Script para descargar e instalar Whisper.cpp para STT local

param(
    [string]$Model = "base",  # tiny, base, small, medium, large
    [string]$Language = "es"
)

$ErrorActionPreference = "Stop"

Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  INSTALADOR DE WHISPER.CPP - STT LOCAL" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Detectar directorio raíz del proyecto
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BinDir = Join-Path $ProjectRoot "bin"
$WhisperDir = Join-Path $BinDir "whisper"
$ModelsDir = Join-Path $ProjectRoot "assets\models\whisper"

Write-Host "IMPORTANTE: Whisper.cpp requiere compilación desde código fuente" -ForegroundColor Yellow
Write-Host ""
Write-Host "Opciones de instalación:" -ForegroundColor Cyan
Write-Host ""
Write-Host "OPCIÓN 1 - Descarga precompilada (Recomendada):" -ForegroundColor Green
Write-Host "  1. Visita: https://github.com/ggerganov/whisper.cpp/releases" -ForegroundColor White
Write-Host "  2. Descarga: whisper-bin-x64.zip (Windows)" -ForegroundColor White
Write-Host "  3. Extrae en: $WhisperDir" -ForegroundColor White
Write-Host ""
Write-Host "OPCIÓN 2 - Compilar desde fuente (Avanzado):" -ForegroundColor Yellow
Write-Host "  Requisitos: Visual Studio 2022, CMake, Git" -ForegroundColor White
Write-Host "  git clone https://github.com/ggerganov/whisper.cpp.git" -ForegroundColor Gray
Write-Host "  cd whisper.cpp" -ForegroundColor Gray
Write-Host "  mkdir build && cd build" -ForegroundColor Gray
Write-Host "  cmake .." -ForegroundColor Gray
Write-Host "  cmake --build . --config Release" -ForegroundColor Gray
Write-Host ""
Write-Host "OPCIÓN 3 - Usar OpenAI STT (más fácil):" -ForegroundColor Cyan
Write-Host "  Configurar provider: 'openai' en ana.config.yaml" -ForegroundColor White
Write-Host "  Requiere: OPENAI_API_KEY en .env" -ForegroundColor White
Write-Host ""

# Preparar directorios
Write-Host "[1/3] Preparando directorios..." -ForegroundColor Yellow
if (-not (Test-Path $WhisperDir)) {
    New-Item -ItemType Directory -Path $WhisperDir -Force | Out-Null
}
if (-not (Test-Path $ModelsDir)) {
    New-Item -ItemType Directory -Path $ModelsDir -Force | Out-Null
}
Write-Host "  ✓ Directorios creados" -ForegroundColor Green

# Descargar modelo automáticamente
Write-Host ""
Write-Host "[2/3] ¿Descargar modelo Whisper '$Model'? (S/N)" -ForegroundColor Yellow
Write-Host "  Tamaños aproximados:" -ForegroundColor Gray
Write-Host "    tiny   = ~75 MB   (rápido, menos preciso)" -ForegroundColor Gray
Write-Host "    base   = ~142 MB  (balance recomendado)" -ForegroundColor Gray
Write-Host "    small  = ~466 MB  (más preciso)" -ForegroundColor Gray
Write-Host "    medium = ~1.5 GB  (muy preciso, lento)" -ForegroundColor Gray
Write-Host "    large  = ~2.9 GB  (máxima precisión)" -ForegroundColor Gray
Write-Host ""

$Response = Read-Host
if ($Response -eq "S" -or $Response -eq "s") {
    $ModelFile = "ggml-$Model.bin"
    $ModelPath = Join-Path $ModelsDir $ModelFile
    $ModelUrl = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/$ModelFile"

    Write-Host ""
    Write-Host "Descargando modelo '$Model'..." -ForegroundColor Yellow
    Write-Host "URL: $ModelUrl" -ForegroundColor Gray

    try {
        $ProgressPreference = 'Continue'
        Invoke-WebRequest -Uri $ModelUrl -OutFile $ModelPath -UseBasicParsing

        $SizeMB = [math]::Round((Get-Item $ModelPath).Length / 1MB, 2)
        Write-Host "  ✓ Descargado: $SizeMB MB" -ForegroundColor Green
        Write-Host "  Ubicación: $ModelPath" -ForegroundColor Cyan
    } catch {
        Write-Host "  ✗ Error al descargar modelo: $_" -ForegroundColor Red
        Write-Host ""
        Write-Host "Modelos disponibles:" -ForegroundColor Yellow
        Write-Host "  tiny, base, small, medium, large, large-v2, large-v3" -ForegroundColor White
        exit 1
    }
} else {
    Write-Host "  ℹ️  Descarga de modelo omitida" -ForegroundColor Gray
}

Write-Host ""
Write-Host "[3/3] Verificando instalación..." -ForegroundColor Yellow

$WhisperExe = Join-Path $WhisperDir "main.exe"
if (Test-Path $WhisperExe) {
    Write-Host "  ✓ Whisper.cpp encontrado: $WhisperExe" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Whisper.cpp NO encontrado en: $WhisperDir" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Descarga manual:" -ForegroundColor Cyan
    Write-Host "  1. Visita: https://github.com/ggerganov/whisper.cpp/releases" -ForegroundColor White
    Write-Host "  2. Descarga el binario precompilado para Windows" -ForegroundColor White
    Write-Host "  3. Extrae main.exe en: $WhisperDir" -ForegroundColor White
    Write-Host ""
}

# Listar modelos descargados
Write-Host ""
Write-Host "Modelos disponibles:" -ForegroundColor Yellow
$Models = Get-ChildItem -Path $ModelsDir -Filter "ggml-*.bin" -ErrorAction SilentlyContinue
if ($Models) {
    foreach ($m in $Models) {
        $SizeMB = [math]::Round($m.Length / 1MB, 2)
        Write-Host "  • $($m.Name) ($SizeMB MB)" -ForegroundColor White
    }
} else {
    Write-Host "  (ninguno descargado)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "  SETUP DE WHISPER COMPLETADO" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""

if (Test-Path $WhisperExe) {
    Write-Host "Configuración para ana.config.yaml:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "stt:" -ForegroundColor White
    Write-Host "  provider: `"whisper`"" -ForegroundColor White
    Write-Host "  whisper:" -ForegroundColor White
    Write-Host "    binary_path: `"./bin/whisper/main.exe`"" -ForegroundColor Cyan
    Write-Host "    model_path: `"./assets/models/whisper/ggml-$Model.bin`"" -ForegroundColor Cyan
    Write-Host "    language: `"$Language`"" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host "ALTERNATIVA MÁS FÁCIL - Usa OpenAI STT:" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "stt:" -ForegroundColor White
    Write-Host "  provider: `"openai`"" -ForegroundColor White
    Write-Host "  openai:" -ForegroundColor White
    Write-Host "    api_key: `"`${OPENAI_API_KEY}`"" -ForegroundColor White
    Write-Host "    model: `"whisper-1`"" -ForegroundColor White
    Write-Host ""
    Write-Host "Esta opción funciona inmediatamente sin compilar nada." -ForegroundColor Gray
    Write-Host ""
}

Write-Host "Para descargar más modelos, ejecuta:" -ForegroundColor Yellow
Write-Host "  .\scripts\install_whisper.ps1 -Model small" -ForegroundColor White
Write-Host ""
