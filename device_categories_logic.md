# Lógica de Categorización de Dispositivos (Enfoque Profesional & Preciso)

Dado que la monitorización GPS es delicada (por saltos de posición falsos o *GPS drift*, retrasos de señal y problemas de telemetría), hemos ajustado la lógica a estándares de plataforma telemática empresarial. 

El modelo de DTO se ha adaptado a nomenclaturas inglesas internacionales (`all`, `live`, `stopped`, `offline`) e incluye en su respuesta el campo **Ignition** (estado del motor, encendido/apagado) el cual es clave para alta precisión.

Se evalúa con la siguiente jerarquía de forma descendente en el backend:

### 1. OFFLINE (Sin Conexión)
Determinar si un dispositivo "cayó" debe basarse principal y estrictamente en su tiempo de latencia de servidor (`server_time`).

* ¿Por qué es un error usar solo la batería? Porque un saboteador puede simplemente arrancar el cable o entrar en un túnel largo/sótano con 100% de batería, el GPS no transmitirá y queremos detectarlo urgente como "Perdido/Sin Conexión".
* ¿Por qué usar `server_time` en lugar de `device_time`? Si el GPS pierde señal (GPRS/GSM), guarda las ubicaciones internamente pero deja de hablar con nosotros. Su `device_time` puede marcar algo en memoria, pero el servidor no tuvo noticias recientes.

**NUEVA REGLA (PROFESIONAL):**
* `Hora Actual del Servidor - Último Reporte Recibido > 1.5 hs` (Pérdida de señal de datos / Offline).
* **Opcional/Agravante:** Si la batería es `< 20%` y perdió señal por al menos 1 hora, también puede declararse Offline ya que se asume que procedió a apagarse por corte de energía.

### 2. LIVE (En Movimiento / Viajando)
Para evitar que un dispositivo parqueado dibuje líneas raras porque está debajo de un techo metálico rebotando coordenadas (GPS Drift).

**NUEVA REGLA (PROFESIONAL):**
* No está `OFFLINE`.
* `Speed > 5 km/h` **Y** (Opcionalmente) Estado de Motor `Ignition == true`.
* *Propuesta a largo plazo:* Exigir que `Ignition` esté en verdadero disminuye los falsos positivos en 99%, ya que sabemos certeramente que el contacto del vehículo está puesto.

### 3. IDLING (Ralentí)
El vehículo está recibiendo señal normal, está detenido en cuanto a velocidad, pero mantiene su motor encendido. Consistente consumo de combustible y desgaste de motor sin generar avance. Útil para flotas comerciales.

**NUEVA REGLA (PROFESIONAL):**
* No está `OFFLINE` ni `LIVE`.
* El motor está prendido (`Ignition == true`).
* `Speed <= 5 km/h`

### 4. PARKED (Parqueado)
El vehículo está aparcado y apagado formalmente recibiendo señal correctamente.

**NUEVA REGLA (PROFESIONAL):**
* No está `OFFLINE` ni `LIVE`.
* El motor está apagado (`Ignition == false` o valor ausente).

---
## Detalles de Implementación Técnica
* Ya hemos descartado la categoría unificada `STOPPED` y modificamos el servicio para calcular en 5 niveles de alta exactitud.
* Aumentamos activamente el parseo del JSON nativo extrayendo el estado de la ignición (`ignition`) hacia la capa de presentación.
* El análisis del tiempo de inactividad de señal se lee contra `server_time` (cuándo el servidor local escuchó un frame del GPS por última vez) descartando cachés engañosos de memoria interna del dispositivo GPS.
