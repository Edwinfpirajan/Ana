# Script para iniciar Ollama y Ana en paralelo

Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "        ANA - INICIANDO OLLAMA Y APLICACIÓN" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

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
}

Write-Host ""

# Iniciar Ollama en una nueva ventana
Write-Host "Iniciando Ollama en nueva ventana..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "ollama serve" -WindowStyle Normal

# Esperar a que Ollama inicie
Write-Host "Esperando a que Ollama se inicie (5 segundos)..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Iniciar Ana
Write-Host ""
Write-Host "Iniciando Ana..." -ForegroundColor Green
Write-Host ""

if (Test-Path .\ana.exe) {
    .\ana.exe
} elseif (Test-Path .\ana) {
    .\ana
} else {
    Write-Host "✗ Ejecutable no encontrado. Compilando..." -ForegroundColor Yellow
    go run .\cmd\ana\main.go
}
