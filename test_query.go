package main

import (
	"fmt"
	"log"


	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/La_Paz",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	query := `
		SELECT 
			d.imei,
			COALESCE(v.placa, '')            AS placa
		FROM devices d
		LEFT JOIN device_installations di ON di.imei = d.imei AND di.removed_at IS NULL
		LEFT JOIN vehicles v ON v.id = di.vehicle_id
		LEFT JOIN models vmo ON vmo.id = v.model_id
		LEFT JOIN makes  vm  ON vm.id  = vmo.make_id
		LEFT JOIN device_last_positions lp ON lp.imei = d.imei
		WHERE 1=1 AND v.user_id = 'abf957c9-c380-4ebd-8fd5-da3c1f7fc323'
	`
	var results []map[string]interface{}
	db.Raw(query).Scan(&results)
	fmt.Printf("Query results count: %d\n", len(results))
	for i, v := range results {
		fmt.Printf("- %v\n", v)
		if i > 5 { break }
	}
}
