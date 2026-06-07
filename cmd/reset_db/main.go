package main

import (
	"log"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/providers"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func main() {
	injector := do.New()
	providers.RegisterDependencies(injector)

	db, err := do.InvokeNamed[*gorm.DB](injector, constants.DB)
	if err != nil {
		log.Fatalf("falib to invoke DB: %v", err)
	}

	// 1. Drop ONLY auxiliary/conflictive tables.
	// CORE DATA (Users, Devices, Positions) is PRESERVED.
	log.Println("Dropping specific auxiliary tables to fix schema conflicts...")

	// GroupDevice: This table had structural issues (missing ID).
	// Dropping it forces recreation with correct schema. Data here is just relationships.
	if err := db.Migrator().DropTable(&entities.GroupDevice{}); err != nil {
		log.Printf("Warning dropping GroupDevice: %v", err)
	} else {
		log.Println("Dropped GroupDevice table (Relationships cleared)")
	}

	// DeviceInstallation: Optional cleanup.
	// If you want to keep installation history, remove this block.
	// Assuming it might have old schema issues too.
	// if err := db.Migrator().DropTable(&entities.DeviceInstallation{}); err != nil {
	// 	log.Printf("Warning dropping DeviceInstallation: %v", err)
	// } else {
	// 	log.Println("Dropped DeviceInstallation table")
	// }

	// DO NOT DROP:
	// - entities.Device (Preserve registered devices)
	// - entities.Position (Preserve location history)
	// - entities.User (Preserve accounts)

	log.Println("---------------------------------------------------------")
	log.Println("Tables 'group_devices' dropped.")
	log.Println("CRITICAL TABLES (Devices, Positions, Users) WERE PRESERVED.")
	log.Println("Now run './motos-api --migrate:run' to:")
	log.Println("1. Re-create 'group_devices' with correct schema.")
	log.Println("2. Apply ALTER TABLE migrations to 'devices' and 'users'.")
	log.Println("---------------------------------------------------------")
}
