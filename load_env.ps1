$envFile = Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) ".env"

if (-Not (Test-Path $envFile)) {
    Write-Error "Archivo .env no encontrado en $envFile"
    exit 1
}

Get-Content $envFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq "" -or $line.StartsWith("#")) {
        return
    }

    if ($line -notmatch "^\s*([^=]+)=(.*)$") {
        return
    }

    $name = $matches[1].Trim()
    $value = $matches[2].Trim()

    if ($value.StartsWith('"') -and $value.EndsWith('"')) {
        $value = $value.Trim('"')
    }

    $envVar = "Env:$name"
    Set-Item -Path $envVar -Value $value
    Write-Host "Variable cargada: $name" -ForegroundColor Cyan
}

Write-Host "Variables de entorno cargadas desde .env" -ForegroundColor Green
