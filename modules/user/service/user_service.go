package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	diRepo "github.com/Caknoooo/go-gin-clean-starter/modules/device_installation/repository"
	groupRepo "github.com/Caknoooo/go-gin-clean-starter/modules/group/repository"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/user/repository"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/helpers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type UserService interface {
	GetUserById(ctx context.Context, userId string) (dto.UserResponse, error)
	GetUserByIdOrEmail(ctx context.Context, identifier string) (dto.UserResponse, error)
	Create(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error)
	Update(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error)
	UpdateMe(ctx context.Context, userId string, req dto.UserMeUpdateRequest) (dto.UserResponse, error)
	ChangeMyPassword(ctx context.Context, userId string, req dto.UserChangePasswordRequest) error
	AdminResetPassword(ctx context.Context, targetUserId string, req dto.AdminResetPasswordRequest) error
	UpdateBlockStatus(ctx context.Context, userId string, isBlocked bool) error
	ChangeStatus(ctx context.Context, userId string) error
	GetInstalledDevices(ctx context.Context, userId string, isAdmin bool) (*dto.UserDevicesAndGroupsResponse, error)
}

type userService struct {
	userRepository         repository.UserRepository
	deviceInstallationRepo diRepo.DeviceInstallationRepository
	groupRepository        groupRepo.GroupRepository
	db                     *gorm.DB
}

func NewUserService(
	userRepo repository.UserRepository,
	deviceInstallationRepo diRepo.DeviceInstallationRepository,
	groupRepository groupRepo.GroupRepository,
	db *gorm.DB,
) UserService {
	return &userService{
		userRepository:         userRepo,
		deviceInstallationRepo: deviceInstallationRepo,
		groupRepository:        groupRepository,
		db:                     db,
	}
}

func (s *userService) GetUserById(ctx context.Context, userId string) (dto.UserResponse, error) {
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:         user.ID.String(),
		Name:       user.Name,
		Username:   user.Username,
		Email:      user.Email,
		TelpNumber: user.TelpNumber,
		Role:       user.Role,
		RoleLiteral: constants.RoleLiteral(user.Role),
		ImageUrl:   user.ImageUrl,
		IsVerified: user.IsVerified,
		IsBlocked:  user.IsBlocked,
		Status:     user.Status,
	}, nil
}

func (s *userService) GetUserByIdOrEmail(ctx context.Context, identifier string) (dto.UserResponse, error) {
	var user entities.User
	var err error

	if strings.Contains(identifier, "@") {
		user, err = s.userRepository.GetUserByEmail(ctx, s.db, identifier)
	} else {
		user, err = s.userRepository.GetUserById(ctx, s.db, identifier)
	}
	if err != nil {
		return dto.UserResponse{}, dto.ErrUserNotFound
	}

	return dto.UserResponse{
		ID:          user.ID.String(),
		Name:        user.Name,
		Username:    user.Username,
		Email:       user.Email,
		TelpNumber:  user.TelpNumber,
		Role:        user.Role,
		RoleLiteral: constants.RoleLiteral(user.Role),
		ImageUrl:    user.ImageUrl,
		IsVerified:  user.IsVerified,
		IsBlocked:   user.IsBlocked,
		Status:      user.Status,
	}, nil
}

func (s *userService) Create(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error) {
	_, isExist, err := s.userRepository.CheckEmail(ctx, s.db, req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return dto.UserResponse{}, err
	}

	if isExist {
		return dto.UserResponse{}, dto.ErrEmailAlreadyExists
	}

	user := entities.User{
		ID:         uuid.New(),
		Name:       req.Name,
		Username:   req.Username,
		Email:      req.Email,
		TelpNumber: req.TelpNumber,
		Password:   req.Password,
		Role:       constants.ENUM_ROLE_USER,
		IsVerified: false,
	}

	createdUser, err := s.userRepository.Register(ctx, s.db, user)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return dto.UserResponse{}, dto.ErrUsernameAlreadyExists
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return dto.UserResponse{}, dto.ErrEmailAlreadyExists
			}
		}
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:          createdUser.ID.String(),
		Name:        createdUser.Name,
		Username:    createdUser.Username,
		Email:       createdUser.Email,
		TelpNumber:  createdUser.TelpNumber,
		Role:        createdUser.Role,
		RoleLiteral: constants.RoleLiteral(createdUser.Role),
		ImageUrl:    createdUser.ImageUrl,
		IsVerified:  createdUser.IsVerified,
		IsBlocked:   createdUser.IsBlocked,
		Status:      createdUser.Status,
	}, nil
}

func (s *userService) Update(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error) {
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.UserUpdateResponse{}, dto.ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.TelpNumber != "" {
		user.TelpNumber = req.TelpNumber
	}

	updatedUser, err := s.userRepository.Update(ctx, s.db, user)
	if err != nil {
		return dto.UserUpdateResponse{}, err
	}

	return dto.UserUpdateResponse{
		ID:         updatedUser.ID.String(),
		Name:       updatedUser.Name,
		TelpNumber: updatedUser.TelpNumber,
		Role:       updatedUser.Role,
		RoleLiteral: constants.RoleLiteral(updatedUser.Role),
		Email:      updatedUser.Email,
		IsVerified: updatedUser.IsVerified,
	}, nil
}

func (s *userService) UpdateMe(ctx context.Context, userId string, req dto.UserMeUpdateRequest) (dto.UserResponse, error) {
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.UserResponse{}, dto.ErrUserNotFound
	}

	// username (nullable)
	if req.Username != nil {
		normalized := *req.Username
		if normalized == "" {
			user.Username = nil
		} else {
			// validar duplicado
			if existing, ok, err := s.userRepository.CheckUsername(ctx, s.db, normalized); err == nil && ok {
				if existing.ID != user.ID {
				return dto.UserResponse{}, dto.ErrUsernameAlreadyExists
				}
			} else if err != nil && err != gorm.ErrRecordNotFound {
				return dto.UserResponse{}, err
			}
			user.Username = &normalized
		}
	}

	if req.Email != "" && req.Email != user.Email {
		if existing, ok, err := s.userRepository.CheckEmail(ctx, s.db, req.Email); err == nil && ok {
			if existing.ID != user.ID {
				return dto.UserResponse{}, dto.ErrEmailAlreadyExists
			}
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return dto.UserResponse{}, err
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.TelpNumber != "" {
		user.TelpNumber = req.TelpNumber
	}
	if req.ImageUrl != nil {
		user.ImageUrl = *req.ImageUrl
	}

	updatedUser, err := s.userRepository.Update(ctx, s.db, user)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:          updatedUser.ID.String(),
		Name:        updatedUser.Name,
		Username:    updatedUser.Username,
		Email:       updatedUser.Email,
		TelpNumber:  updatedUser.TelpNumber,
		Role:        updatedUser.Role,
		RoleLiteral: constants.RoleLiteral(updatedUser.Role),
		ImageUrl:    updatedUser.ImageUrl,
		IsVerified:  updatedUser.IsVerified,
		IsBlocked:   updatedUser.IsBlocked,
		Status:      updatedUser.Status,
	}, nil
}

func (s *userService) ChangeMyPassword(ctx context.Context, userId string, req dto.UserChangePasswordRequest) error {
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.ErrUserNotFound
	}

	ok, err := helpers.CheckPassword(user.Password, []byte(req.CurrentPassword))
	if err != nil || !ok {
		return dto.ErrCurrentPasswordInvalid
	}

	hashed, err := helpers.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.userRepository.UpdatePassword(ctx, s.db, userId, hashed)
}

func (s *userService) AdminResetPassword(ctx context.Context, targetUserId string, req dto.AdminResetPasswordRequest) error {
	// validar que el usuario exista
	if _, err := s.userRepository.GetUserById(ctx, s.db, targetUserId); err != nil {
		return dto.ErrUserNotFound
	}

	hashed, err := helpers.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.userRepository.UpdatePassword(ctx, s.db, targetUserId, hashed)
}

func (s *userService) UpdateBlockStatus(ctx context.Context, userId string, isBlocked bool) error {
	// Optional: Check if user exists first?
	// The repo update will fail or return nil status if not found, but standard GORM update by ID doesn't error on record not found unless filtered.
	// We can verify user exists if we want specific error.
	_, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.ErrUserNotFound
	}

	return s.userRepository.UpdateBlockStatus(ctx, s.db, userId, isBlocked)
}

func (s *userService) ChangeStatus(ctx context.Context, userId string) error {
	user, err := s.userRepository.GetUserById(ctx, s.db, userId)
	if err != nil {
		return dto.ErrUserNotFound
	}
	// Toggling status
	newStatus := !user.Status
	return s.userRepository.UpdateStatus(ctx, s.db, userId, newStatus)
}

func (s *userService) GetInstalledDevices(ctx context.Context, userId string, isAdmin bool) (*dto.UserDevicesAndGroupsResponse, error) {
	id, err := uuid.Parse(userId)
	if err != nil {
		return nil, err
	}

	// 1. Get ALL installed devices for user (Individual)
	var installations []entities.DeviceInstallation
	if isAdmin {
		err = s.db.WithContext(ctx).
			Joins("JOIN vehicles ON vehicles.id = device_installations.vehicle_id").
			Where("device_installations.status = ?", true).
			Where("device_installations.removed_at IS NULL").
			Preload("Device").
			Preload("Vehicle").
			Preload("Vehicle.User").
			Preload("Vehicle.Model").
			Preload("Vehicle.Model.Make").
			Find(&installations).Error
	} else {
		installations, err = s.deviceInstallationRepo.FindAllByUserID(ctx, id)
	}
	if err != nil {
		return nil, err
	}

	// Helper to map installation entity to DTO
	mapInstallationToDTO := func(inst entities.DeviceInstallation) dto.UserInstalledDeviceResponse {
		makeName := ""
		modelName := ""
		if inst.Vehicle != nil && inst.Vehicle.Model != nil {
			modelName = inst.Vehicle.Model.ModelName
			if inst.Vehicle.Model.Make != nil {
				makeName = inst.Vehicle.Model.Make.MakeName
			}
		}

		deviceModel := ""
		if inst.Device != nil {
			deviceModel = inst.Device.Model
		}

		response := dto.UserInstalledDeviceResponse{
			InstallationID: inst.InstallationID,
			InstalledAt:    inst.InstalledAt,
			Status:         inst.Status,
		}

		if inst.Device != nil {
			response.Device = dto.DeviceInfo{
				IMEI:       inst.Device.IMEI,
				DeviceType: deviceModel,
			}
		}

		if inst.Vehicle != nil {
			response.Vehicle = dto.VehicleInfo{
				ID:          inst.Vehicle.ID,
				Placa:       inst.Vehicle.Placa,
				Description: inst.Vehicle.Description,
				Year:        inst.Vehicle.Year,
				Model:       modelName,
				Make:        makeName,
				Color:       inst.Vehicle.Color,
			}
		}
		return response
	}

	// Convert individual installations
	installedDevices := make([]dto.UserInstalledDeviceResponse, 0, len(installations))
	for _, inst := range installations {
		installedDevices = append(installedDevices, mapInstallationToDTO(inst))
	}

	// 2. Get Groups
	deviceGroups := make([]dto.GroupDeviceInfo, 0)
	var groups []entities.Group
	if isAdmin {
		groups, err = s.groupRepository.FindAll(ctx)
	} else {
		groups, err = s.groupRepository.FindAllByUserID(ctx, id)
	}
	if err != nil {
		// Return partial result if groups fail
		return &dto.UserDevicesAndGroupsResponse{
			InstalledDevices: installedDevices,
			DeviceGroups:     deviceGroups,
		}, nil
	}

	if len(groups) == 0 {
		return &dto.UserDevicesAndGroupsResponse{
			InstalledDevices: installedDevices,
			DeviceGroups:     deviceGroups,
		}, nil
	}

	// 3. Batch fetch Group Devices (Optimization: 1 Query)
	groupIDs := make([]uuid.UUID, len(groups))
	for i, g := range groups {
		groupIDs[i] = g.ID
	}

	var allGroupDevices []entities.GroupDevice
	err = s.db.WithContext(ctx).
		Where("group_id IN ? AND status = ?", groupIDs, true).
		Find(&allGroupDevices).Error
	if err != nil {
		return nil, err
	}

	// 4. Collect IMEIs from groups
	imeiSet := make(map[string]struct{})
	groupToIMEIs := make(map[uuid.UUID][]string)

	for _, gd := range allGroupDevices {
		if gd.DeviceIMEI != "" { // Assuming DeviceIMEI is the FK field
			imeiSet[gd.DeviceIMEI] = struct{}{}
			groupToIMEIs[gd.GroupID] = append(groupToIMEIs[gd.GroupID], gd.DeviceIMEI)
		}
	}

	var imeis []string
	for imei := range imeiSet {
		imeis = append(imeis, imei)
	}

	// 5. Batch fetch active installations for these IMEIs belonging to user (Optimization: 1 Query)
	// We need to verify these devices are installed in user's vehicles
	var groupInstallations []entities.DeviceInstallation
	if len(imeis) > 0 {
		err = s.db.WithContext(ctx).
			Joins("JOIN vehicles ON vehicles.id = device_installations.vehicle_id").
			Where("device_installations.imei IN ?", imeis).
			Where("device_installations.removed_at IS NULL AND device_installations.status = ?", true).
			Preload("Device").
			Preload("Vehicle").
			Preload("Vehicle.Model").
			Preload("Vehicle.Model.Make").
			Find(&groupInstallations).Error
		if err != nil {
			return nil, err
		}
	}

	// 6. Map Installations by IMEI for quick lookup
	installationsByIMEI := make(map[string][]dto.UserInstalledDeviceResponse)
	for _, inst := range groupInstallations {
		if inst.Device != nil {
			dtoInst := mapInstallationToDTO(inst)
			installationsByIMEI[inst.Device.IMEI] = append(installationsByIMEI[inst.Device.IMEI], dtoInst)
		}
	}

	// 7. Build Response
	for _, group := range groups {
		groupDTO := dto.GroupDeviceInfo{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			CreatedAt:   group.CreatedAt,
			Devices:     make([]dto.UserInstalledDeviceResponse, 0),
		}

		// Get devices for this group
		if deviceIMEIs, ok := groupToIMEIs[group.ID]; ok {
			for _, imei := range deviceIMEIs {
				// Get installations for this device
				if insts, found := installationsByIMEI[imei]; found {
					groupDTO.Devices = append(groupDTO.Devices, insts...)
				}
			}
		}

		deviceGroups = append(deviceGroups, groupDTO)
	}

	return &dto.UserDevicesAndGroupsResponse{
		InstalledDevices: installedDevices,
		DeviceGroups:     deviceGroups,
	}, nil
}
