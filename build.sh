#!/bin/bash
export PATH=/c/msys64/mingw64/bin:/c/msys64/usr/bin:$PATH
export CGO_ENABLED=1
export CC=gcc.exe
export PKG_CONFIG_PATH=/c/msys64/mingw64/lib/pkgconfig

echo '═══════════════════════════════════════════'
echo 'Building Ana with name activation filter'
echo '═══════════════════════════════════════════'
echo ''

cd /c/Users/Ferchando/Documents/ana
/c/Program\ Files/Go/bin/go.exe build -tags portaudio -o ana.exe ./cmd/ana/main.go

if [ $? -eq 0 ]; then
    echo ''
    echo 'Build successful!'
    echo ''
else
    echo ''
    echo 'Build failed!'
    exit 1
fi
