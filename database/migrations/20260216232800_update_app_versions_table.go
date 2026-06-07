package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260216232800_update_app_versions_table", Up20260216232800UpdateAppVersionsTable, Down20260216232800UpdateAppVersionsTable)
}

func Up20260216232800UpdateAppVersionsTable(db *gorm.DB) error {
	// Change version_code from int to varchar(10)
	if err := db.Exec("ALTER TABLE app_versions ALTER COLUMN version_code TYPE varchar(10)").Error; err != nil {
		return err
	}

	// Increase URL fields to 500 characters
	if err := db.Exec("ALTER TABLE app_versions ALTER COLUMN url_playstore TYPE varchar(500)").Error; err != nil {
		return err
	}

	if err := db.Exec("ALTER TABLE app_versions ALTER COLUMN url_applestore TYPE varchar(500)").Error; err != nil {
		return err
	}

	// Drop the state column
	if err := db.Exec("ALTER TABLE app_versions DROP COLUMN IF EXISTS state").Error; err != nil {
		return err
	}

	return nil
}

func Down20260216232800UpdateAppVersionsTable(db *gorm.DB) error {
	// Revert version_code to int
	if err := db.Exec("ALTER TABLE app_versions ALTER COLUMN version_code TYPE int USING version_code::integer").Error; err != nil {
		return err
	}

	// Revert URL fields to 255 characters
	if err := db.Exec("ALTER TABLE app_versions ALTER COLUMN url_playstore TYPE varchar(255)").Error; err != nil {
		return err
	}

	if err := db.Exec("ALTER TABLE app_versions ALTER COLUMN url_applestore TYPE varchar(255)").Error; err != nil {
		return err
	}

	// Add back the state column
	if err := db.Exec("ALTER TABLE app_versions ADD COLUMN state boolean DEFAULT true").Error; err != nil {
		return err
	}

	return nil
}
