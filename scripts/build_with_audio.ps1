$profilePath = Split-Path -Parent $MyInvocation.MyCommand.Path
$env:INCLUDE = "$env:INCLUDE;C:\tools\vcpkg\installed\x64-windows\include"
$env:LIB = "$env:LIB;C:\tools\vcpkg\installed\x64-windows\lib"
$env:CGO_CFLAGS = "-IC:\tools\vcpkg\installed\x64-windows\include"
$env:CGO_LDFLAGS = "-LC:\tools\vcpkg\installed\x64-windows\lib -lportaudio"
$env:CGO_ENABLED = "1"
$env:PATH = "C:\msys64\mingw64\bin;C:\msys64\usr\bin;$env:PATH"
$env:PKG_CONFIG_PATH = "C:\tools\vcpkg\installed\x64-windows\lib\pkgconfig"

Write-Host "Building Ana with PortAudio support..."
cd "$profilePath\.."
go build -tags portaudio ./cmd/ana
if ($LASTEXITCODE -ne 0) {
  Write-Error "Build failed."
  exit $LASTEXITCODE
}
Write-Host "Build succeeded."
