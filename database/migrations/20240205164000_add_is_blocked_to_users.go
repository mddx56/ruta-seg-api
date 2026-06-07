package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20240205164000_add_is_blocked_to_users", Up20240205164000, Down20240205164000)
}

func Up20240205164000(db *gorm.DB) error {
	return db.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS is_blocked boolean DEFAULT false").Error
}

func Down20240205164000(db *gorm.DB) error {
	return db.Exec("ALTER TABLE users DROP COLUMN IF EXISTS is_blocked").Error
}
