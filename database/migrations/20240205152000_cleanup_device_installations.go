package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20240205152000_cleanup_device_installations", Up20240205152000, Down20240205152000)
}

func Up20240205152000(db *gorm.DB) error {
	// 1. Drop old foreign keys
	db.Exec("ALTER TABLE device_installations DROP CONSTRAINT IF EXISTS fk_device_installations_device")
	db.Exec("ALTER TABLE device_installations DROP CONSTRAINT IF EXISTS fk_devices_installations")

	// 2. Drop the device_id column if it exists (since we use imei now)
	db.Exec("ALTER TABLE device_installations DROP COLUMN IF EXISTS device_id")

	// 3. Add foreign key to imei column
	return db.Exec("ALTER TABLE device_installations ADD CONSTRAINT fk_device_installations_device FOREIGN KEY (imei) REFERENCES devices(imei) ON UPDATE CASCADE ON DELETE CASCADE").Error
}

func Down20240205152000(db *gorm.DB) error {
	// Reversed: add device_id back if needed, but usually we don't rollback this far in dev
	return nil
}
