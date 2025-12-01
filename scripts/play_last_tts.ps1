$pattern = "ana_play_*"
$tempDir = [IO.Path]::GetTempPath()

$files = Get-ChildItem -Path $tempDir -Filter "$pattern*.{wav,mp3}" -File -ErrorAction SilentlyContinue
if (-not $files) {
    Write-Host "No se encontró ningún archivo de TTS reciente en $tempDir" -ForegroundColor Yellow
    exit 1
}

$latestFile = $files | Sort-Object LastWriteTime -Descending | Select-Object -First 1
Write-Host "Reproduciendo $($latestFile.FullName)" -ForegroundColor Green

try {
    Add-Type -AssemblyName presentationCore
    $player = New-Object system.windows.media.mediaplayer
    $player.open($latestFile.FullName)
    while ($player.NaturalDuration -eq [TimeSpan]::Zero) {
        Start-Sleep -Milliseconds 50
    }
    $player.Play()
    while ($player.Position -lt $player.NaturalDuration) {
        Start-Sleep -Milliseconds 100
    }
} catch {
    Write-Host "Error al reproducir: $_" -ForegroundColor Red
    exit 1
}
