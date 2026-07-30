package seeds

import (
	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"gorm.io/gorm"
)

// FineTypeSeeder carga el catálogo fijo de tipos de multa. Es idempotente: solo
// inserta los códigos que todavía no existen.
func FineTypeSeeder(db *gorm.DB) error {
	fineTypes := []entities.FineType{
		{Code: "OVERTAKING", Name: "Adelantamiento entre micros", DefaultAmount: 50, Severity: "WARNING"},
		{Code: "LAP_TIME", Name: "Tiempo de vuelta fuera de lo permitido", DefaultAmount: 30, Severity: "WARNING"},
		{Code: "PROLONGED_STOP", Name: "Parada prolongada", DefaultAmount: 20, Severity: "WARNING"},
		{Code: "SPEEDING", Name: "Exceso de velocidad", DefaultAmount: 100, Severity: "CRITICAL"},
	}

	for _, ft := range fineTypes {
		var count int64
		if err := db.Model(&entities.FineType{}).Where("code = ?", ft.Code).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Create(&ft).Error; err != nil {
			return err
		}
	}

	return nil
}
