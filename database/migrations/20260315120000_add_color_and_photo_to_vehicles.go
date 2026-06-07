package migrations

import (
	"github.com/Caknoooo/go-gin-clean-starter/database"
	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260315120000_add_color_and_photo_to_vehicles", Up20260315120000AddColorAndPhotoToVehicles, Down20260315120000AddColorAndPhotoToVehicles)
}

func Up20260315120000AddColorAndPhotoToVehicles(db *gorm.DB) error {
	// Add columns to "vehicles" table (using the Vehicle model metadata)
	if !db.Migrator().HasColumn(&entities.Vehicle{}, "color") {
		if err := db.Migrator().AddColumn(&entities.Vehicle{}, "Color"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(&entities.Vehicle{}, "photo_url") {
		if err := db.Migrator().AddColumn(&entities.Vehicle{}, "PhotoURL"); err != nil {
			return err
		}
	}
	return nil
}

func Down20260315120000AddColorAndPhotoToVehicles(db *gorm.DB) error {
	if db.Migrator().HasColumn(&entities.Vehicle{}, "color") {
		if err := db.Migrator().DropColumn(&entities.Vehicle{}, "Color"); err != nil {
			return err
		}
	}
	if db.Migrator().HasColumn(&entities.Vehicle{}, "photo_url") {
		if err := db.Migrator().DropColumn(&entities.Vehicle{}, "PhotoURL"); err != nil {
			return err
		}
	}
	return nil
}

