
# Script para compilar version Linux (Lubuntu/Ubuntu) desde Windows
Write-Host "Compilando para Linux (amd64)..."

$original_goos = $env:GOOS
$original_goarch = $env:GOARCH

$env:GOOS = "linux"
$env:GOARCH = "amd64"

# Crear directorio bin si no existe
if (!(Test-Path -Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

go build -o bin/motos-api ./cmd/main.go
go build -o bin/reset-db ./cmd/reset_db/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build exitoso: bin/motos-api y bin/reset-db"
} else {
    Write-Host "❌ Error en el build"
}

# Restaurar variables de entorno (opcional)
$env:GOOS = $original_goos
$env:GOARCH = $original_goarch
