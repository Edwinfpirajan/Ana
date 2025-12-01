# ═══════════════════════════════════════════════════════════════════════════════
#                    INSTALADOR DE PIPER TTS (Windows)
# ═══════════════════════════════════════════════════════════════════════════════
# Script para descargar e instalar Piper TTS localmente

param(
    [string]$Version = "2023.11.14-2",
    [string]$Arch = "amd64"  # amd64 o arm64
)

$ErrorActionPreference = "Stop"

Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  INSTALADOR DE PIPER TTS - 100% LOCAL" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Detectar directorio raíz del proyecto
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BinDir = Join-Path $ProjectRoot "bin"
$PiperDir = Join-Path $BinDir "piper"

Write-Host "[1/4] Preparando directorios..." -ForegroundColor Yellow
if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
}
if (Test-Path $PiperDir) {
    Write-Host "  ⚠️  Piper ya existe, eliminando versión anterior..." -ForegroundColor Yellow
    Remove-Item -Path $PiperDir -Recurse -Force
}
New-Item -ItemType Directory -Path $PiperDir -Force | Out-Null

# Construir URL de descarga
$FileName = "piper_windows_$Arch.zip"
$DownloadUrl = "https://github.com/rhasspy/piper/releases/download/$Version/$FileName"
$ZipPath = Join-Path $BinDir $FileName

Write-Host "[2/4] Descargando Piper desde GitHub..." -ForegroundColor Yellow
Write-Host "  URL: $DownloadUrl" -ForegroundColor Gray
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
    Write-Host "  ✓ Descarga completada" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Error al descargar Piper: $_" -ForegroundColor Red
    exit 1
}

Write-Host "[3/4] Extrayendo archivos..." -ForegroundColor Yellow
try {
    Expand-Archive -Path $ZipPath -DestinationPath $PiperDir -Force
    Write-Host "  ✓ Extracción completada" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Error al extraer: $_" -ForegroundColor Red
    exit 1
}

Write-Host "[4/4] Limpiando archivos temporales..." -ForegroundColor Yellow
Remove-Item -Path $ZipPath -Force
Write-Host "  ✓ Limpieza completada" -ForegroundColor Green

# Verificar instalación
$PiperExe = Join-Path $PiperDir "piper.exe"
if (Test-Path $PiperExe) {
    Write-Host ""
    Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
    Write-Host "  ✓ PIPER INSTALADO EXITOSAMENTE" -ForegroundColor Green
    Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Green
    Write-Host ""
    Write-Host "Ubicación: $PiperExe" -ForegroundColor Cyan
    Write-Host ""

    # Probar versión
    try {
        $VersionOutput = & $PiperExe --version 2>&1
        Write-Host "Versión instalada: $VersionOutput" -ForegroundColor Cyan
    } catch {
        Write-Host "Nota: No se pudo verificar la versión, pero el ejecutable existe" -ForegroundColor Yellow
    }

    Write-Host ""
    Write-Host "Próximos pasos:" -ForegroundColor Yellow
    Write-Host "  1. Ejecuta: .\scripts\download_voices.ps1" -ForegroundColor White
    Write-Host "  2. Configura ana.config.yaml con:" -ForegroundColor White
    Write-Host "     binary_path: `"./bin/piper/piper.exe`"" -ForegroundColor Gray
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "✗ Error: No se encontró piper.exe después de la instalación" -ForegroundColor Red
    exit 1
}
