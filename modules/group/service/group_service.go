package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	deviceInstallationRepo "github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/group/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/group/repository"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type GroupService interface {
	Create(ctx context.Context, req dto.GroupCreateRequest) (dto.GroupResponse, error)
	Update(ctx context.Context, req dto.GroupUpdateRequest) (dto.GroupResponse, error)
	ChangeStatus(ctx context.Context, id string) error
	FindAll(ctx context.Context) ([]dto.GroupResponse, error)
	FindAllByUserID(ctx context.Context, userID string) ([]dto.GroupResponse, error)
	AssignDevice(ctx context.Context, req dto.GroupAssignDeviceRequest, userID string) error
	RemoveDevice(ctx context.Context, req dto.GroupRemoveDeviceRequest) error
}

type groupService struct {
	repo                   repository.GroupRepository
	deviceInstallationRepo deviceInstallationRepo.DeviceInstallationRepository
}

func NewGroupService(injector *do.Injector) (GroupService, error) {
	repo := do.MustInvoke[repository.GroupRepository](injector)
	deviceInstallationRepo := do.MustInvoke[deviceInstallationRepo.DeviceInstallationRepository](injector)
	return &groupService{
		repo:                   repo,
		deviceInstallationRepo: deviceInstallationRepo,
	}, nil
}

func (s *groupService) Create(ctx context.Context, req dto.GroupCreateRequest) (dto.GroupResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return dto.GroupResponse{}, err
	}

	group := entities.Group{
		Name:   req.Name,
		UserID: userID,
	}
	if req.Description != "" {
		group.Description = &req.Description
	}

	if err := s.repo.Create(ctx, &group); err != nil {
		return dto.GroupResponse{}, err
	}

	return s.mapEntityToDto(group), nil
}

func (s *groupService) Update(ctx context.Context, req dto.GroupUpdateRequest) (dto.GroupResponse, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return dto.GroupResponse{}, err
	}

	group, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.GroupResponse{}, err
	}

	if req.Name != "" {
		group.Name = req.Name
	}
	if req.Description != "" {
		group.Description = &req.Description
	}

	if err := s.repo.Update(ctx, &group); err != nil {
		return dto.GroupResponse{}, err
	}

	return s.mapEntityToDto(group), nil
}

func (s *groupService) ChangeStatus(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return s.repo.ChangeStatus(ctx, uid)
}

func (s *groupService) FindAll(ctx context.Context) ([]dto.GroupResponse, error) {
	groups, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.GroupResponse, 0)
	for _, group := range groups {
		responses = append(responses, s.mapEntityToDto(group))
	}
	return responses, nil
}

func (s *groupService) FindAllByUserID(ctx context.Context, userID string) ([]dto.GroupResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	groups, err := s.repo.FindAllByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.GroupResponse, 0)
	for _, group := range groups {
		responses = append(responses, s.mapEntityToDto(group))
	}
	return responses, nil
}

func (s *groupService) AssignDevice(ctx context.Context, req dto.GroupAssignDeviceRequest, userID string) error {
	// 1. Verify group exists
	_, err := s.repo.FindByID(ctx, req.GroupID)
	if err != nil {
		return err
	}

	// 2. Verify device has an active installation (is associated with a vehicle)
	_, err = s.deviceInstallationRepo.FindActiveByIMEI(ctx, req.DeviceIMEI)
	if err != nil {
		return fmt.Errorf("el dispositivo %s no tiene una instalación activa (no está asociado a un vehículo)", req.DeviceIMEI)
	}

	// 3. Check if device is already in group
	assigned, err := s.repo.IsDeviceAssignedToGroup(ctx, req.GroupID, req.DeviceIMEI)
	if err != nil {
		return err
	}
	if assigned {
		return errors.New("el dispositivo ya está asignado a este grupo")
	}

	// 4. Assign device
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	groupDevice := entities.GroupDevice{
		GroupID:    req.GroupID,
		DeviceIMEI: req.DeviceIMEI,
		AssignedBy: uid,
	}

	return s.repo.AssignDevice(ctx, &groupDevice)
}

func (s *groupService) RemoveDevice(ctx context.Context, req dto.GroupRemoveDeviceRequest) error {
	return s.repo.RemoveDevice(ctx, req.GroupID, req.DeviceIMEI)
}

func (s *groupService) mapEntityToDto(group entities.Group) dto.GroupResponse {
	return dto.GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		UserID:      group.UserID,
	}
}
