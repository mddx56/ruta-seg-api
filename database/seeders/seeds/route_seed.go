package seeds

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"gorm.io/gorm"
)

// RouteSeeder carga el trazado fijo de la Línea 18 (primer anillo) desde un GeoJSON
// pre-generado y lo guarda como Route.Geometry. Es idempotente: si la ruta ya existe, no hace nada.
func RouteSeeder(db *gorm.DB) error {
	const routeName = "Línea 18 - Primer Anillo"

	var count int64
	if err := db.Model(&entities.Route{}).Where("name = ?", routeName).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	geoJSONFile, err := os.Open("./database/seeders/data/linea-18.geojson")
	if err != nil {
		return err
	}
	defer geoJSONFile.Close()

	geoJSONBytes, err := io.ReadAll(geoJSONFile)
	if err != nil {
		return err
	}

	if !json.Valid(geoJSONBytes) {
		return errors.New("linea-18.geojson no contiene un JSON válido")
	}
	geometry := string(geoJSONBytes)

	var admin entities.User
	if err := db.Where("role = ?", constants.ENUM_ROLE_ADMIN).First(&admin).Error; err != nil {
		return err
	}

	route := entities.Route{
		Name:        routeName,
		Geometry:    &geometry,
		CreatedByID: admin.ID,
	}

	return db.Create(&route).Error
}
