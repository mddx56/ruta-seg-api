package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260402200000_add_vehicle_position_indexes", Up20260402200000AddVehiclePositionIndexes, Down20260402200000AddVehiclePositionIndexes)
}

func Up20260402200000AddVehiclePositionIndexes(db *gorm.DB) error {
	sqls := []string{
		// Composite index para queries de historial por vehículo:
		// WHERE device_id = X AND device_time >= T1 AND device_time < T2
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_device_time
			ON positions(device_id, device_time ASC)`,

		// Partial index para obtener instalaciones activas por vehículo rápidamente:
		// WHERE vehicle_id = X AND removed_at IS NULL
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_device_installations_vehicle_active
			ON device_installations(vehicle_id)
			WHERE removed_at IS NULL`,

		// Index para queries de rango temporal de instalaciones:
		// WHERE vehicle_id = X AND installed_at < end AND (removed_at IS NULL OR removed_at > start)
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_device_installations_vehicle_range
			ON device_installations(vehicle_id, installed_at, removed_at)`,
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func Down20260402200000AddVehiclePositionIndexes(db *gorm.DB) error {
	sqls := []string{
		`DROP INDEX CONCURRENTLY IF EXISTS idx_positions_device_time`,
		`DROP INDEX CONCURRENTLY IF EXISTS idx_device_installations_vehicle_active`,
		`DROP INDEX CONCURRENTLY IF EXISTS idx_device_installations_vehicle_range`,
	}
	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}
