#!/bin/bash
# ============================================================================
# Ana Streamer Build Script for Linux/macOS
# ============================================================================
# This script compiles Ana Streamer with PortAudio support
# Requirements: Go, GCC/Clang, and PortAudio development libraries

set -e

echo ""
echo "============================================================================"
echo " Ana Streamer - Linux/macOS Build Script"
echo "============================================================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed or not in PATH"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

echo "[1/3] Checking Go installation..."
GO_VERSION=$(go version | awk '{print $3}')
echo "       Go version: $GO_VERSION"
echo ""

# Check for PortAudio development files
echo "[2/3] Checking PortAudio installation..."
if pkg-config --exists portaudio-2.0 2>/dev/null; then
    echo "       PortAudio: Found"
    PORTAUDIO_VERSION=$(pkg-config --modversion portaudio-2.0)
    echo "       Version: $PORTAUDIO_VERSION"
elif [ -f "/usr/include/portaudio.h" ] || [ -f "/usr/local/include/portaudio.h" ]; then
    echo "       PortAudio: Found (headers only)"
else
    echo "       WARNING: PortAudio development files not found"
    echo "       Install with:"
    echo "         Ubuntu/Debian: sudo apt-get install portaudio19-dev"
    echo "         macOS: brew install portaudio"
    echo "         Fedora/RHEL: sudo dnf install portaudio-devel"
    echo ""
fi
echo ""

echo "[3/3] Building Ana Streamer..."
echo "       Command: go build -tags portaudio -o ana ./cmd/ana/main.go"
echo ""

# Set build environment
export CGO_ENABLED=1

# Run the actual build
if go build -tags portaudio -o ana ./cmd/ana/main.go; then
    # Get file size
    SIZE=$(ls -lh ana | awk '{print $5}')
    echo ""
    echo "============================================================================"
    echo " BUILD SUCCESSFUL"
    echo "============================================================================"
    echo " Output: ana ($SIZE)"
    echo ""
    echo " Run Ana with: ./ana"
    echo "============================================================================"
    echo ""
else
    echo ""
    echo "============================================================================"
    echo " BUILD FAILED"
    echo "============================================================================"
    echo ""
    echo " Troubleshooting:"
    echo " 1. Ensure PortAudio dev libs are installed"
    echo " 2. Install PortAudio development files:"
    echo "    Ubuntu/Debian: sudo apt-get install portaudio19-dev"
    echo "    macOS: brew install portaudio"
    echo " 3. Check CGO compilation: go env CGO_ENABLED"
    echo "============================================================================"
    exit 1
fi
