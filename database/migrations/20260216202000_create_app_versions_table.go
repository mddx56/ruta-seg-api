package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260216202000_create_app_versions_table", Up20260216202000CreateAppVersionsTable, Down20260216202000CreateAppVersionsTable)
}

func Up20260216202000CreateAppVersionsTable(db *gorm.DB) error {
	return db.AutoMigrate(&entities.AppVersion{})
}

func Down20260216202000CreateAppVersionsTable(db *gorm.DB) error {
	return db.Migrator().DropTable(&entities.AppVersion{})
}
