# Script para ejecutar Ana con variables de entorno cargadas (PowerShell)

# Cargar variables de entorno desde .env
if (Test-Path .\.env) {
    Get-Content .\.env | ForEach-Object {
        if ($_ -match '^\s*([^=]+)=(.*)$') {
            $key = $matches[1].Trim()
            $value = $matches[2].Trim()
            [System.Environment]::SetEnvironmentVariable($key, $value, "Process")
        }
    }
    Write-Host "✓ Variables de entorno cargadas" -ForegroundColor Green
} else {
    Write-Host "✗ Archivo .env no encontrado" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "                    ANA - INICIANDO" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "✓ Ejecutando Ana..." -ForegroundColor Green
Write-Host ""

# Usar el ejecutable compilado con soporte de audio
if (Test-Path .\ana.exe) {
    .\ana.exe
} elseif (Test-Path .\ana) {
    .\ana
} else {
    Write-Host "✗ Ejecutable no encontrado. Compilando..." -ForegroundColor Yellow
    go run .\cmd\ana\main.go
}
