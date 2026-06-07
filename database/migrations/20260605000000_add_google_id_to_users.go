package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260605000000_add_google_id_to_users", Up20260605000000, Down20260605000000)
}

func Up20260605000000(db *gorm.DB) error {
	return db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id varchar(255) UNIQUE").Error
}

func Down20260605000000(db *gorm.DB) error {
	return db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS google_id").Error
}
