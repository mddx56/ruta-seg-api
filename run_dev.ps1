Write-Host "Generating Swagger docs..." -ForegroundColor Cyan
swag init -g cmd/main.go

if ($LASTEXITCODE -ne 0) {
    Write-Warning "Swagger generation failed! Continuing..."
}

Write-Host "Starting application..." -ForegroundColor Green
go run cmd/main.go
