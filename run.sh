#!/bin/bash
# Script para ejecutar Ana con variables de entorno cargadas

set -a
source .env
set +a

echo "═══════════════════════════════════════════════════════════════"
echo "                    ANA - INICIANDO"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "✓ Variables de entorno cargadas"
echo "✓ Ejecutando Ana..."
echo ""

# Usar el ejecutable compilado con soporte de audio
if [ -f ./ana.exe ]; then
    ./ana.exe
elif [ -f ./ana ]; then
    ./ana
else
    echo "✗ Ejecutable no encontrado. Compilando..."
    /c/Program\ Files/Go/bin/go run ./cmd/ana/main.go
fi
