# Auth API — Documentación para SPA

**Base URL:** `https://apiscz.rutasegurascz.com`

Todas las respuestas siguen esta estructura:
```json
{
  "status": true,
  "message": "descripción",
  "data": { ... }
}
```

---

## Estructura reutilizable: UserResponse

Devuelta en Login y Refresh dentro del campo `data.user`:

```json
{
  "id":           "uuid-string",
  "name":         "Juan Pérez",
  "username":     "juanp123",
  "email":        "juan@email.com",
  "telp_number":  "591-70000000",
  "role":         "user",
  "role_literal": "Usuario",
  "image_url":    "https://...",
  "is_verified":  true,
  "is_blocked":   false,
  "status":       true
}
```

---

## 1. Signup — Registro de usuario

```
POST /api/auth/signup
```

> Rate limit: 3 requests / minuto por IP.

### Payload
```json
{
  "name":        "Juan Pérez",
  "email":       "juan@email.com",
  "password":    "mipass123",
  "username":    "juanp123",
  "telp_number": "591-70000000"
}
```

| Campo        | Tipo   | Requerido | Validación         |
|--------------|--------|-----------|--------------------|
| `name`       | string | ✅        | min 2, max 100     |
| `email`      | string | ✅        | formato email      |
| `password`   | string | ✅        | min 4 caracteres   |
| `username`   | string | ❌        | min 4, max 20      |
| `telp_number`| string | ❌        | min 7, max 20      |

### ✅ 200 OK
```json
{
  "status": true,
  "message": "usuario creado correctamente",
  "data": {
    "id":           "uuid",
    "name":         "Juan Pérez",
    "username":     "juanp123",
    "email":        "juan@email.com",
    "telp_number":  "591-70000000",
    "role":         "user",
    "role_literal": "Usuario",
    "image_url":    "",
    "is_verified":  false,
    "is_blocked":   false,
    "status":       true
  }
}
```

### ❌ 400 Bad Request
```json
{ "status": false, "message": "fallo crear usuario", "data": "email ya existe" }
{ "status": false, "message": "fallo crear usuario", "data": "username ya existe" }
{ "status": false, "message": "fallo obtener datos del body", "data": "..." }
```

### ❌ 429 Too Many Requests
```json
{ "status": false, "message": "fallo crear usuario", "data": "demasiados intentos" }
```

---

## 2. Login

```
POST /api/auth/login
```

> Acepta login por **email** o por **username** (enviar uno de los dos).
> Bloqueado tras 5 intentos fallidos por 15 minutos.

### Payload — con email
```json
{
  "email":    "juan@email.com",
  "password": "mipass123"
}
```

### Payload — con username
```json
{
  "username": "juanp123",
  "password": "mipass123"
}
```

### ✅ 200 OK
```json
{
  "status": true,
  "message": "sesion iniciada correctamente",
  "data": {
    "access_token":  "eyJhbGci...",
    "refresh_token": "base64string...",
    "user": { ... },
    "app": {
      "app_id":                "com.rutasegura.app",
      "version_name":          "1.0.0",
      "version_code":          1,
      "url_playstore":         "https://...",
      "url_applestore":        "https://...",
      "fecha_release":         "2025-01-01T00:00:00Z",
      "mini_supported_version": 1,
      "is_force_update":       false,
      "plataform":             "android",
      "created_at":            "2025-01-01T00:00:00Z",
      "updated_at":            "2025-01-01T00:00:00Z"
    }
  }
}
```

> `app` puede ser `null` si no hay versión configurada.

### ❌ 401 Unauthorized
```json
{ "status": false, "message": "fallo iniciar sesion", "data": "credenciales invalidas" }
{ "status": false, "message": "fallo iniciar sesion", "data": "usuario no encontrado" }
```

### ❌ 423 Locked
```json
{ "status": false, "message": "fallo iniciar sesion", "data": "usuario sin acceso, comuniquese con el administrador" }
{ "status": false, "message": "fallo iniciar sesion", "data": "cuenta desactivada, contacte al administrador" }
```

### ❌ 429 Too Many Requests
```json
{ "status": false, "message": "fallo iniciar sesion", "data": "demasiados intentos fallidos, intente de nuevo en 15 minutos" }
```

---

## 3. Refresh Token

```
POST /api/auth/refresh
```

### Payload
```json
{
  "refresh_token": "base64string..."
}
```

### ✅ 200 OK
```json
{
  "status": true,
  "message": "renovacion exitosa de token",
  "data": {
    "access_token":  "eyJhbGci...",
    "refresh_token": "nuevoBase64string...",
    "user": { ... },
    "app": { ... }
  }
}
```

> El refresh token anterior queda **invalidado**. Guardar el nuevo.

### ❌ 401 Unauthorized
```json
{ "status": false, "message": "fallo al renovar token", "data": "refresh token no encontrado" }
```

---

## 4. Logout

```
POST /api/auth/logout
Authorization: Bearer <access_token>
```

> No requiere body. Invalida todos los refresh tokens del usuario.

### ✅ 200 OK
```json
{
  "status": true,
  "message": "cerrar sesion exitoso",
  "data": null
}
```

### ❌ 400 Bad Request
```json
{ "status": false, "message": "fallo al cerrar sesion", "data": "..." }
```

### ❌ 401 Unauthorized — token inválido o ausente
```json
{ "status": false, "message": "token no valido", "data": null }
```

---

## 5. Enviar Código de Verificación

```
POST /api/auth/send-verification-email
```

> Envía un código OTP de 4 dígitos al email. Expira en 10 minutos.

### Payload
```json
{
  "email": "juan@email.com"
}
```

### ✅ 200 OK
```json
{
  "status": true,
  "message": "Email de verificación enviado",
  "data": null
}
```

### ❌ 400 Bad Request
```json
{ "status": false, "message": "fallo procesar solicitud", "data": "email no encontrado" }
{ "status": false, "message": "fallo procesar solicitud", "data": "cuenta ya verificada" }
```

---

## 6. Verificar Email con OTP

```
POST /api/auth/verify-email
```

> El usuario recibe un código de 4 dígitos en su email (ej: `7392`).

### Payload
```json
{
  "email": "juan@email.com",
  "otp":   "7392"
}
```

| Campo   | Tipo   | Validación          |
|---------|--------|---------------------|
| `email` | string | requerido, formato email |
| `otp`   | string | requerido, exactamente 4 caracteres |

### ✅ 200 OK
```json
{
  "status": true,
  "message": "email verificado correctamente",
  "data": {
    "email":       "juan@email.com",
    "is_verified": true
  }
}
```

### ❌ 400 Bad Request
```json
{ "status": false, "message": "fallo verificar email", "data": "token invalid" }
{ "status": false, "message": "fallo verificar email", "data": "usuario no encontrado" }
```

---

## 7. Solicitar Reset de Contraseña

```
POST /api/auth/send-password-reset
```

### Payload
```json
{
  "email": "juan@email.com"
}
```

### ✅ 200 OK
```json
{
  "status": true,
  "message": "envio exitoso de solicitud de restablecimiento de contraseña",
  "data": null
}
```

### ❌ 400 Bad Request
```json
{ "status": false, "message": "fallo al enviar solicitud de restablecimiento de contraseña", "data": "email no encontrado" }
```

---

## 8. Resetear Contraseña

```
POST /api/auth/reset-password
```

> El token llega al email. Copiar y enviar aquí junto a la nueva contraseña.

### Payload
```json
{
  "token":        "eyJhbGci...",
  "new_password": "nuevaPass123"
}
```

| Campo          | Tipo   | Validación       |
|----------------|--------|------------------|
| `token`        | string | requerido        |
| `new_password` | string | min 4 caracteres |

### ✅ 200 OK
```json
{
  "status": true,
  "message": "restablecimiento exitoso de contraseña",
  "data": null
}
```

### ❌ 400 Bad Request
```json
{ "status": false, "message": "fallo al restablecer contraseña", "data": "token de restablecimiento de contraseña invalido" }
{ "status": false, "message": "fallo al restablecer contraseña", "data": "usuario no encontrado" }
```

---

## Manejo del Access Token en el SPA

Adjuntar en todas las requests protegidas:

```http
Authorization: Bearer eyJhbGci...
```

**Vida útil:** 4 horas  
**Refresh token:** 7 días

### Flujo recomendado
1. Guardar ambos tokens en `localStorage` o `sessionStorage`
2. En cada request, verificar si el `access_token` expiró (decodificar JWT, campo `exp`)
3. Si expiró → llamar `/api/auth/refresh` con el `refresh_token`
4. Guardar el nuevo par de tokens
5. Si el refresh también falló (401) → redirigir al login
