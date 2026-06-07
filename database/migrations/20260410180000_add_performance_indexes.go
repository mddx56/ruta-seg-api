package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260410180000_add_performance_indexes", Up20260410180000AddPerformanceIndexes, Down20260410180000AddPerformanceIndexes)
}

// Up20260410180000AddPerformanceIndexes agrega todos los índices necesarios para
// reducir la complejidad del endpoint GET /api/devices/categories de
// O(N × P × log P) → O(N log N) con el índice compuesto crítico en positions.
//
// ┌─────────────────────────────────────────────────────────────────────────┐
// │  ÍNDICE                              │ MEJORA                           │
// ├─────────────────────────────────────────────────────────────────────────┤
// │  positions(device_id, server_time)   │ LATERAL LIMIT 1 → O(1) por dev  │
// │  vehicles(user_id)                   │ WHERE user_id=? → O(log V)       │
// │  vehicles(placa)                     │ ORDER BY placa → sin filesort    │
// │  device_installations(status,imei)   │ WHERE status=true → partial idx  │
// └─────────────────────────────────────────────────────────────────────────┘
func Up20260410180000AddPerformanceIndexes(db *gorm.DB) error {
	sqls := []string{
		// ── 🔴 CRÍTICO: Composite index en positions ────────────────────────
		// Convierte el LATERAL JOIN para obtener la última posición de cada device
		// de O(P_i × log P_i) → O(1) usando index-only scan.
		// La query: WHERE device_id = ? ORDER BY server_time DESC LIMIT 1
		// pasará a usar un "Index Scan Backward" sin ningún filesort.
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_device_server_time
			ON positions (device_id, server_time DESC)`,

		// ── 🟡 ALTO: Index en vehicles.user_id ─────────────────────────────
		// El filtro WHERE v.user_id = ? actualmente hace un Seq Scan de toda
		// la tabla vehicles. Con este índice → O(log V).
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_user_id
			ON vehicles (user_id)`,

		// ── 🟡 ALTO: Index en vehicles.placa ───────────────────────────────
		// El ORDER BY v.placa ASC necesita un filesort sobre N filas.
		// Placa ya tiene UNIQUE (que implica un índice B-Tree), GORM lo crea
		// automáticamente. Este índice es redundante para ordenamiento pero se
		// deja explícito por claridad. Si ya existe, IF NOT EXISTS previene error.
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_placa
			ON vehicles (placa ASC)`,

		// ── 🟢 MEDIO: Partial index en device_installations activas ────────
		// WHERE di.removed_at IS NULL AND di.status = true es el filtro base
		// de TODAS las queries de instalaciones activas.
		// Un partial index solo indexa las filas que cumplen la condición WHERE,
		// haciéndolo mucho más pequeño y rápido que un índice completo.
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_di_active_imei
			ON device_installations (imei)
			WHERE removed_at IS NULL AND status = true`,

		// ── 🟢 MEDIO: Index en device_installations.vehicle_id activas ─────
		// El JOIN vehicles v ON v.id = di.vehicle_id necesita este índice
		// para el lookup desde installations hacia vehicles.
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_di_active_vehicle_id
			ON device_installations (vehicle_id)
			WHERE removed_at IS NULL AND status = true`,

		// ── ℹ️  BONUS: vehicle_models y vehicle_makes ───────────────────────
		// Los LEFT JOINs de modelo y marca ya usan sus PKs (uuid), PostgreSQL
		// crea automáticamente los B-Tree index en PKs. No se necesitan extras.

		// ── ℹ️  BONUS: positions.device_id ya tiene índice simple (entidad) ─
		// El nuevo índice compuesto (device_id, server_time DESC) reemplaza
		// funcionalmente al simple para la query LATERAL. PostgreSQL usará el
		// más selectivo automáticamente.
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func Down20260410180000AddPerformanceIndexes(db *gorm.DB) error {
	sqls := []string{
		`DROP INDEX CONCURRENTLY IF EXISTS idx_positions_device_server_time`,
		`DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_user_id`,
		`DROP INDEX CONCURRENTLY IF EXISTS idx_vehicles_placa`,
		`DROP INDEX CONCURRENTLY IF EXISTS idx_di_active_imei`,
		`DROP INDEX CONCURRENTLY IF EXISTS idx_di_active_vehicle_id`,
	}
	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}
