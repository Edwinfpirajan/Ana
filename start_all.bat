@echo off
REM Script para iniciar Ollama y Ana en paralelo

setlocal enabledelayedexpansion

echo ═══════════════════════════════════════════════════════════════
echo         ANA - INICIANDO OLLAMA Y APLICACIÓN
echo ═══════════════════════════════════════════════════════════════
echo.

REM Cargar variables desde .env
if exist .env (
    for /f "usebackq delims==" %%A in (.env) do (
        set %%A
    )
    echo ✓ Variables de entorno cargadas
)

echo.
echo Iniciando Ollama en nueva ventana...
start "Ollama" cmd /k "ollama serve"

REM Esperar a que Ollama inicie
echo Esperando a que Ollama se inicie (5 segundos)...
timeout /t 5 /nobreak

echo.
echo Iniciando Ana...
echo.

if exist ana.exe (
    ana.exe
) else if exist ana (
    ana
) else (
    echo ✗ Ejecutable no encontrado. Compilando...
    go run .\cmd\ana\main.go
)
