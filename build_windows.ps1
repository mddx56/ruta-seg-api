
# Script para compilar version Windows desde Windows
Write-Host "Compilando para Windows (amd64)..."

$original_goos = $env:GOOS
$original_goarch = $env:GOARCH

$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Crear directorio bin si no existe
if (!(Test-Path -Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

go build -o bin/motos-api.exe ./cmd/main.go
go build -o bin/reset-db.exe ./cmd/reset_db/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build exitoso: bin/motos-api.exe y bin/reset-db.exe"
} else {
    Write-Host "❌ Error en el build"
}

# Restaurar variables
$env:GOOS = $original_goos
$env:GOARCH = $original_goarch
