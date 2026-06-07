package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260410190000_seed_device_last_positions", Up20260410190000SeedDeviceLastPositions, Down20260410190000SeedDeviceLastPositions)
}

// Up20260410190000SeedDeviceLastPositions pre-llena device_last_positions
// con los datos históricos ya existentes en la tabla positions.
// Después de esta migración, futuros INSERT en positions harán UPSERT 
// automáticamente vía el position_controller.
func Up20260410190000SeedDeviceLastPositions(db *gorm.DB) error {
	// Llenar device_last_positions con la última posición de cada device
	// usando DISTINCT ON de PostgreSQL (mucho más eficiente que GROUP BY + subquery)
	sql := `
		INSERT INTO device_last_positions (imei, latitude, longitude, speed, course, device_time, server_time, attributes, updated_at)
		SELECT DISTINCT ON (device_id)
			device_id,
			latitude,
			longitude,
			speed,
			course,
			device_time,
			server_time,
			attributes,
			NOW()
		FROM positions
		ORDER BY device_id, server_time DESC
		ON CONFLICT (imei) DO UPDATE SET
			latitude    = EXCLUDED.latitude,
			longitude   = EXCLUDED.longitude,
			speed       = EXCLUDED.speed,
			course      = EXCLUDED.course,
			device_time = EXCLUDED.device_time,
			server_time = EXCLUDED.server_time,
			attributes  = EXCLUDED.attributes,
			updated_at  = NOW()
	`
	return db.Exec(sql).Error
}

func Down20260410190000SeedDeviceLastPositions(db *gorm.DB) error {
	return db.Exec("TRUNCATE TABLE device_last_positions").Error
}
