package service

import (
	"context"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/app_version/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/app_version/repository"
	"github.com/samber/do"
)

type AppVersionService interface {
	Create(ctx context.Context, req dto.AppVersionRequest) (dto.AppVersionResponse, error)
	GetLatestVersion(ctx context.Context) (dto.AppVersionResponse, error)
	GetAll(ctx context.Context) ([]dto.AppVersionResponse, error)
	GetById(ctx context.Context, appId int) (dto.AppVersionResponse, error)
	Update(ctx context.Context, req dto.AppVersionRequest, appId int) (dto.AppVersionResponse, error)
	ChangeStatus(ctx context.Context, appId int) error
	Delete(ctx context.Context, appId int) error
}

type appVersionService struct {
	appVersionRepo repository.AppVersionRepository
}

func NewAppVersionService(injector *do.Injector) (AppVersionService, error) {
	appVersionRepo := do.MustInvoke[repository.AppVersionRepository](injector)
	return &appVersionService{
		appVersionRepo: appVersionRepo,
	}, nil
}

func (s *appVersionService) Create(ctx context.Context, req dto.AppVersionRequest) (dto.AppVersionResponse, error) {
	releaseDate, err := time.Parse("2006-01-02", req.FechaRelease)
	if err != nil {
		return dto.AppVersionResponse{}, err
	}

	appVersion := entities.AppVersion{
		VersionName:          req.VersionName,
		VersionCode:          req.VersionCode,
		UrlPlaystore:         req.UrlPlaystore,
		UrlApplestore:        req.UrlApplestore,
		FechaRelease:         releaseDate,
		MiniSupportedVersion: req.MiniSupportedVersion,
		IsForceUpdate:        req.IsForceUpdate,
		Plataform:            req.Plataform,
	}

	res, err := s.appVersionRepo.Create(ctx, nil, appVersion)
	if err != nil {
		return dto.AppVersionResponse{}, err
	}

	return s.mapEntityToDto(res), nil
}

func (s *appVersionService) GetLatestVersion(ctx context.Context) (dto.AppVersionResponse, error) {
	res, err := s.appVersionRepo.GetLatestVersion(ctx, nil)
	if err != nil {
		return dto.AppVersionResponse{}, err
	}

	return s.mapEntityToDto(res), nil
}

func (s *appVersionService) GetAll(ctx context.Context) ([]dto.AppVersionResponse, error) {
	res, err := s.appVersionRepo.GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	var appVersions []dto.AppVersionResponse
	for _, av := range res {
		appVersions = append(appVersions, s.mapEntityToDto(av))
	}

	return appVersions, nil
}

func (s *appVersionService) GetById(ctx context.Context, appId int) (dto.AppVersionResponse, error) {
	res, err := s.appVersionRepo.GetById(ctx, nil, appId)
	if err != nil {
		return dto.AppVersionResponse{}, err
	}

	return s.mapEntityToDto(res), nil
}

func (s *appVersionService) Update(ctx context.Context, req dto.AppVersionRequest, appId int) (dto.AppVersionResponse, error) {
	releaseDate, err := time.Parse("2006-01-02", req.FechaRelease)
	if err != nil {
		return dto.AppVersionResponse{}, err
	}

	appVersion, err := s.appVersionRepo.GetById(ctx, nil, appId)
	if err != nil {
		return dto.AppVersionResponse{}, err
	}

	appVersion.VersionName = req.VersionName
	appVersion.VersionCode = req.VersionCode
	appVersion.UrlPlaystore = req.UrlPlaystore
	appVersion.UrlApplestore = req.UrlApplestore
	appVersion.FechaRelease = releaseDate
	appVersion.MiniSupportedVersion = req.MiniSupportedVersion
	appVersion.IsForceUpdate = req.IsForceUpdate
	appVersion.Plataform = req.Plataform

	res, err := s.appVersionRepo.Update(ctx, nil, appVersion)
	if err != nil {
		return dto.AppVersionResponse{}, err
	}

	return s.mapEntityToDto(res), nil
}

func (s *appVersionService) Delete(ctx context.Context, appId int) error {
	return s.appVersionRepo.Delete(ctx, nil, appId)
}

func (s *appVersionService) ChangeStatus(ctx context.Context, appId int) error {
	return s.appVersionRepo.ChangeStatus(ctx, nil, appId)
}

func (s *appVersionService) mapEntityToDto(appVersion entities.AppVersion) dto.AppVersionResponse {
	return dto.AppVersionResponse{
		AppId:                appVersion.AppId,
		VersionName:          appVersion.VersionName,
		VersionCode:          appVersion.VersionCode,
		UrlPlaystore:         appVersion.UrlPlaystore,
		UrlApplestore:        appVersion.UrlApplestore,
		FechaRelease:         appVersion.FechaRelease,
		MiniSupportedVersion: appVersion.MiniSupportedVersion,
		IsForceUpdate:        appVersion.IsForceUpdate,
		Plataform:            appVersion.Plataform,
		CreatedAt:            appVersion.CreatedAt,
		UpdatedAt:            appVersion.UpdatedAt,
	}
}
