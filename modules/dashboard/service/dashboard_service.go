package service

import (
	"context"

	"github.com/Caknoooo/go-gin-clean-starter/modules/dashboard/dto"
	"gorm.io/gorm"
)

type DashboardService interface {
	GetStats(ctx context.Context) (dto.DashboardStatsResponse, error)
}

type dashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) DashboardService {
	return &dashboardService{
		db: db,
	}
}

func (s *dashboardService) GetStats(ctx context.Context) (dto.DashboardStatsResponse, error) {
	var stats dto.DashboardStatsResponse

	// Total de usuarios
	if err := s.db.WithContext(ctx).Table("users").Count(&stats.TotalUsers).Error; err != nil {
		return stats, err
	}

	// Usuarios suspendidos (bloqueados)
	if err := s.db.WithContext(ctx).Table("users").Where("is_blocked = ?", true).Count(&stats.SuspendedUsers).Error; err != nil {
		return stats, err
	}

	// Usuarios con instalaciones activas (usando DISTINCT)
	if err := s.db.WithContext(ctx).
		Table("device_installations").
		Select("COUNT(DISTINCT user_creation_id)").
		Where("removed_at IS NULL AND user_creation_id IS NOT NULL").
		Scan(&stats.UsersWithInstallations).Error; err != nil {
		return stats, err
	}

	// Total de dispositivos
	if err := s.db.WithContext(ctx).Table("devices").Count(&stats.TotalDevices).Error; err != nil {
		return stats, err
	}

	// Dispositivos con instalación activa
	if err := s.db.WithContext(ctx).
		Table("devices").
		Joins("INNER JOIN device_installations ON devices.imei = device_installations.imei").
		Where("device_installations.removed_at IS NULL").
		Count(&stats.DevicesWithInstallation).Error; err != nil {
		return stats, err
	}

	// Dispositivos sin instalación
	stats.DevicesWithoutInstallation = stats.TotalDevices - stats.DevicesWithInstallation

	// Total de vehículos
	if err := s.db.WithContext(ctx).Table("vehicles").Count(&stats.TotalVehicles).Error; err != nil {
		return stats, err
	}

	// Total de grupos
	if err := s.db.WithContext(ctx).Table("groups").Count(&stats.TotalGroups).Error; err != nil {
		return stats, err
	}

	return stats, nil
}
