package database

import (
	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS micros").Error; err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&entities.Migration{},
		&entities.User{},
		&entities.RefreshToken{},
		&entities.Position{},
		&entities.AlarmType{},
		&entities.AlarmRule{},
		&entities.AlarmIncident{},
		&entities.Geofence{},
		&entities.GeofencePoint{},
		&entities.Make{},
		&entities.VehicleType{},
		&entities.Model{},
		&entities.Vehicle{},
		&entities.Device{},
		&entities.DeviceInstallation{},
		&entities.GroupDevice{},
		&entities.LogSocket{},
		&entities.AppVersion{},
		&entities.DeviceLastPosition{}, // Caché de última posición por device
		&entities.Route{},
		&entities.RouteStop{},
		&entities.VehicleRoute{},
		&entities.Lap{},
		&entities.RouteFare{},
		&entities.LapCharge{},
		&entities.FineType{},
		&entities.Fine{},
		&entities.UserDeviceToken{},
		&entities.Notification{},
	); err != nil {
		return err
	}

	manager := NewMigrationManager(db)
	return manager.Run()
}
