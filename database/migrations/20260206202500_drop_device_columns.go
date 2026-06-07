package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260206202500_drop_device_columns", Up20260206202500, Down20260206202500)
}

func Up20260206202500(db *gorm.DB) error {
	// PASO 0: Asegurar que la tabla group_devices tenga la estructura correcta
	// BORRAMOS y RECREAMOS para asegurar integridad, ya que el usuario confirmó que esto era aceptable.
	if err := db.Exec("DROP TABLE IF EXISTS group_devices").Error; err != nil {
		return err
	}

	createTableQuery := `
		CREATE TABLE group_devices (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			group_id uuid NOT NULL,
			device_imei varchar(20) NOT NULL,
			assigned_by uuid NOT NULL,
			created_at timestamptz,
			updated_at timestamptz,
			status boolean DEFAULT true
		);
	`
	if err := db.Exec(createTableQuery).Error; err != nil {
		return err
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_group_devices_group_id ON group_devices(group_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_group_devices_device_imei ON group_devices(device_imei)")

	// PASO 1: VERIFICAR SI EXISTEN DATOS PARA MIGRAR
	// Verificamos si la columna 'group_id' existe en la tabla 'devices'.
	var exists bool
	checkColumnQuery := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name='devices' AND column_name='group_id'
		);
	`
	if err := db.Raw(checkColumnQuery).Scan(&exists).Error; err != nil {
		return err
	}

	if exists {
		// La columna existe, hacemos el backup
		backfillQuery := `
			INSERT INTO group_devices (id, group_id, device_imei, assigned_by, created_at, updated_at, status)
			SELECT 
				gen_random_uuid(), 
				d.group_id, 
				d.imei, 
				g.user_id, 
				NOW(), 
				NOW(), 
				true
			FROM devices d
			JOIN groups g ON g.id = d.group_id
			WHERE d.group_id IS NOT NULL 
			AND NOT EXISTS (
				SELECT 1 FROM group_devices gd 
				WHERE gd.device_imei = d.imei AND gd.group_id = d.group_id
			);
		`
		if err := db.Exec(backfillQuery).Error; err != nil {
			return err
		}

		// Y borramos las columnas viejas
		if err := db.Exec("ALTER TABLE devices DROP COLUMN IF EXISTS user_id").Error; err != nil {
			return err
		}
		if err := db.Exec("ALTER TABLE devices DROP COLUMN IF EXISTS group_id").Error; err != nil {
			return err
		}
	} else {
		// La columna group_id ya no existe en devices, significa que los datos ya fueron migrados o borrados.
		// No hacemos nada mas.
	}

	return nil
}

func Down20260206202500(db *gorm.DB) error {
	// Revertir: Volver a agregar las columnas (sin constraints de FK para simplificar rollback)
	if err := db.Exec("ALTER TABLE devices ADD COLUMN IF NOT EXISTS user_id uuid").Error; err != nil {
		return err
	}
	return db.Exec("ALTER TABLE devices ADD COLUMN IF NOT EXISTS group_id uuid").Error
}
