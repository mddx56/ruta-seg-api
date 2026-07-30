package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/fine/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/fine/repository"
	notificationservice "github.com/Caknoooo/go-gin-clean-starter/modules/notification/service"
	"github.com/google/uuid"
	"github.com/samber/do"
)

// GenerateFineInput es el contrato interno que usa el motor de reglas (Lap, y más
// adelante adelantamiento/parada/velocidad) para registrar una multa automática.
type GenerateFineInput struct {
	VehicleID       uuid.UUID
	FineTypeCode    string
	LapID           *uuid.UUID
	AlarmIncidentID *uuid.UUID
	Amount          *float64 // opcional; si es nil, se usa FineType.DefaultAmount
	Latitude        float64
	Longitude       float64
	OccurredAt      time.Time
	Notes           *string
}

type FineService interface {
	GenerateFine(ctx context.Context, input GenerateFineInput) (entities.Fine, error)
	Void(ctx context.Context, id uuid.UUID, notes *string) error
	FindAll(ctx context.Context) ([]dto.FineResponse, error)
	FindAllMine(ctx context.Context, ownerUserID uuid.UUID) ([]dto.FineResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (dto.FineResponse, error)
	FindAllTypes(ctx context.Context) ([]dto.FineTypeResponse, error)
}

type fineService struct {
	repo                repository.FineRepository
	typeRepo            repository.FineTypeRepository
	notificationService notificationservice.NotificationService
}

func NewFineService(injector *do.Injector) (FineService, error) {
	repo := do.MustInvoke[repository.FineRepository](injector)
	typeRepo := do.MustInvoke[repository.FineTypeRepository](injector)
	notificationSvc, _ := do.Invoke[notificationservice.NotificationService](injector)
	return &fineService{
		repo:                repo,
		typeRepo:            typeRepo,
		notificationService: notificationSvc,
	}, nil
}

func (s *fineService) GenerateFine(ctx context.Context, input GenerateFineInput) (entities.Fine, error) {
	fineType, err := s.typeRepo.FindByCode(ctx, input.FineTypeCode)
	if err != nil {
		return entities.Fine{}, err
	}

	amount := fineType.DefaultAmount
	if input.Amount != nil {
		amount = *input.Amount
	}

	fine := entities.Fine{
		VehicleID:       input.VehicleID,
		FineTypeID:      fineType.ID,
		LapID:           input.LapID,
		AlarmIncidentID: input.AlarmIncidentID,
		Amount:          amount,
		FineStatus:      "PENDING",
		Latitude:        input.Latitude,
		Longitude:       input.Longitude,
		OccurredAt:      input.OccurredAt,
		Notes:           input.Notes,
	}

	if err := s.repo.Create(ctx, &fine); err != nil {
		return entities.Fine{}, err
	}

	fine.FineType = fineType
	s.notifyOwner(ctx, fine)

	return fine, nil
}

// notifyOwner avisa al dueño del micro que se generó una multa nueva (RF-12): la
// persiste en su historial de notificaciones, la manda por el WebSocket privado y,
// si tiene un dispositivo registrado, dispara el push. Best-effort: nunca falla GenerateFine.
func (s *fineService) notifyOwner(ctx context.Context, fine entities.Fine) {
	if s.notificationService == nil {
		return
	}

	ownerID, err := s.repo.FindVehicleOwnerID(ctx, fine.VehicleID)
	if err != nil {
		return
	}

	title := "Nueva multa: " + fine.FineType.Name
	message := fmt.Sprintf("Se generó una multa de Bs. %.2f a tu micro por %s.", fine.Amount, fine.FineType.Name)
	data := map[string]interface{}{
		"fine_id":    fine.ID.String(),
		"vehicle_id": fine.VehicleID.String(),
		"amount":     fine.Amount,
		"fine_type":  fine.FineType.Code,
	}

	if err := s.notificationService.Notify(ctx, ownerID, "FINE_GENERATED", title, message, data); err != nil {
		log.Printf("[fine] error notificando multa %s al dueño %s: %v", fine.ID, ownerID, err)
	}
}

func (s *fineService) Void(ctx context.Context, id uuid.UUID, notes *string) error {
	return s.repo.UpdateStatus(ctx, id, "VOIDED", notes)
}

func (s *fineService) FindAll(ctx context.Context) ([]dto.FineResponse, error) {
	fines, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.FineResponse, 0, len(fines))
	for _, f := range fines {
		responses = append(responses, toResponse(f))
	}

	return responses, nil
}

func (s *fineService) FindAllMine(ctx context.Context, ownerUserID uuid.UUID) ([]dto.FineResponse, error) {
	fines, err := s.repo.FindAllByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.FineResponse, 0, len(fines))
	for _, f := range fines {
		responses = append(responses, toResponse(f))
	}

	return responses, nil
}

func (s *fineService) FindByID(ctx context.Context, id uuid.UUID) (dto.FineResponse, error) {
	fine, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.FineResponse{}, err
	}

	return toResponse(fine), nil
}

func (s *fineService) FindAllTypes(ctx context.Context) ([]dto.FineTypeResponse, error) {
	fineTypes, err := s.typeRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.FineTypeResponse, 0, len(fineTypes))
	for _, ft := range fineTypes {
		responses = append(responses, toTypeResponse(ft))
	}

	return responses, nil
}

func toTypeResponse(ft entities.FineType) dto.FineTypeResponse {
	return dto.FineTypeResponse{
		ID:            ft.ID,
		Code:          ft.Code,
		Name:          ft.Name,
		DefaultAmount: ft.DefaultAmount,
		Severity:      ft.Severity,
	}
}

func toResponse(f entities.Fine) dto.FineResponse {
	return dto.FineResponse{
		ID:              f.ID,
		VehicleID:       f.VehicleID,
		FineType:        toTypeResponse(f.FineType),
		LapID:           f.LapID,
		AlarmIncidentID: f.AlarmIncidentID,
		Amount:          f.Amount,
		FineStatus:      f.FineStatus,
		Latitude:        f.Latitude,
		Longitude:       f.Longitude,
		OccurredAt:      f.OccurredAt,
		Notes:           f.Notes,
		CreatedAt:       f.CreatedAt,
	}
}
