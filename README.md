# 🚀 Comandos Rápidos

## 1️⃣ Base de Datos (Setup)

```bash
# Instalar dependencias
go mod download

# Migraciones
go run cmd/main.go --migrate:run

# Seeds (Datos de prueba)
go run cmd/main.go --seed
```

## 2️⃣ Desarrollo (Sin compilar)

```bash
go run cmd/main.go
```

## 3️⃣ Generar Ejecutable (Build)

**Linux / Mac (Optimizado):**

```bash
go build -ldflags="-s -w" -o bin/app cmd/main.go
```

**Windows (Optimizado):**

```powershell
go build -ldflags="-s -w" -o bin/app.exe cmd/main.go
```

## 4️⃣ Compilación Cruzada (Desde Windows para Linux)

Útil si desarrollas en Windows pero subes a servidor Linux.

```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags="-s -w" -o bin/app-linux cmd/main.go
```

```
```
