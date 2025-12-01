@echo off
REM Script para ejecutar Ana con variables de entorno cargadas (CMD)

setlocal enabledelayedexpansion

echo ═══════════════════════════════════════════════════════════════
echo                     ANA - INICIANDO
echo ═══════════════════════════════════════════════════════════════
echo.

REM Cargar variables desde .env
if exist .env (
    for /f "usebackq delims==" %%A in (.env) do (
        set %%A
    )
    echo ✓ Variables de entorno cargadas
) else (
    echo ✗ Archivo .env no encontrado
    exit /b 1
)

echo ✓ Ejecutando Ana...
echo.

REM Usar el ejecutable compilado con soporte de audio
if exist ana.exe (
    ana.exe
) else if exist ana (
    ana
) else (
    echo ✗ Ejecutable no encontrado. Compilando...
    go run .\cmd\ana\main.go
)

pause
