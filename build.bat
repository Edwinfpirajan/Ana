@echo off
REM ============================================================================
REM Ana Streamer Build Script for Windows
REM ============================================================================
REM This script compiles Ana Streamer with PortAudio support on Windows
REM Requirements: Go, GCC (from MinGW/MSYS2), and PortAudio dev libraries

setlocal enabledelayedexpansion

echo.
echo ============================================================================
echo  Ana Streamer - Windows Build Script
echo ============================================================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    exit /b 1
)

echo [1/3] Checking Go installation...
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo       Go version: !GO_VERSION!
echo.

echo [2/3] Setting up build environment...
REM Set required environment variables for CGO and PortAudio
set CGO_ENABLED=1
set CC=gcc
echo       CGO_ENABLED=1
echo       CC=gcc
echo.

echo [3/3] Building Ana Streamer...
echo       Command: go build -tags portaudio -o ana.exe ./cmd/ana/main.go
echo.

REM Run the actual build
go build -tags portaudio -o ana.exe ./cmd/ana/main.go

if %ERRORLEVEL% EQU 0 (
    REM Get file size
    for %%F in (ana.exe) do set SIZE=%%~zF
    echo.
    echo ============================================================================
    echo  BUILD SUCCESSFUL
    echo ============================================================================
    echo  Output: ana.exe (!SIZE! bytes^)
    echo.
    echo  Run Ana with: ana.exe
    echo ============================================================================
    echo.
) else (
    echo.
    echo ============================================================================
    echo  BUILD FAILED
    echo ============================================================================
    echo  Error code: %ERRORLEVEL%
    echo.
    echo  Troubleshooting:
    echo  1. Ensure GCC is installed: gcc --version
    echo  2. Check PortAudio headers are available
    echo  3. Run from MinGW/MSYS2 shell if CGO fails
    echo ============================================================================
    exit /b %ERRORLEVEL%
)

endlocal
