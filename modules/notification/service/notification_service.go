package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/Caknoooo/go-gin-clean-starter/database/entities"
	"github.com/Caknoooo/go-gin-clean-starter/modules/notification/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/notification/repository"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/fcm"
	providerWS "github.com/Caknoooo/go-gin-clean-starter/providers/websocket"
	"github.com/google/uuid"
	"github.com/samber/do"
)

// PushSender abstrae el envío de push (satisfecho por *fcm.Client), para poder probar
// NotificationService sin llamar a la API real de FCM.
type PushSender interface {
	Send(ctx context.Context, msg fcm.Message) error
}

type NotificationService interface {
	RegisterDeviceToken(ctx context.Context, userID uuid.UUID, token, platform string) error
	UnregisterDeviceToken(ctx context.Context, token string) error
	FindAllMine(ctx context.Context, userID uuid.UUID) ([]dto.NotificationResponse, error)
	MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	// Notify persiste una Notification, la envía por el WebSocket privado existente y,
	// si hay tokens FCM registrados, dispara el push. Es el punto único que deben usar
	// otros módulos (Fine, y a futuro Lap/otros) para avisarle algo a un usuario.
	Notify(ctx context.Context, userID uuid.UUID, notifType, title, message string, data map[string]interface{}) error
}

type notificationService struct {
	repo      repository.NotificationRepository
	tokenRepo repository.DeviceTokenRepository
	wsService providerWS.WebsocketService
	push      PushSender
}

func NewNotificationService(injector *do.Injector) (NotificationService, error) {
	repo := do.MustInvoke[repository.NotificationRepository](injector)
	tokenRepo := do.MustInvoke[repository.DeviceTokenRepository](injector)
	wsSvc, _ := do.Invoke[providerWS.WebsocketService](injector)
	push, _ := do.Invoke[PushSender](injector)
	return &notificationService{
		repo:      repo,
		tokenRepo: tokenRepo,
		wsService: wsSvc,
		push:      push,
	}, nil
}

func (s *notificationService) RegisterDeviceToken(ctx context.Context, userID uuid.UUID, token, platform string) error {
	return s.tokenRepo.Upsert(ctx, &entities.UserDeviceToken{
		UserID:   userID,
		Token:    token,
		Platform: platform,
	})
}

func (s *notificationService) UnregisterDeviceToken(ctx context.Context, token string) error {
	return s.tokenRepo.DeleteByToken(ctx, token)
}

func (s *notificationService) FindAllMine(ctx context.Context, userID uuid.UUID) ([]dto.NotificationResponse, error) {
	notifications, err := s.repo.FindAllByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		responses = append(responses, toResponse(n))
	}
	return responses, nil
}

func (s *notificationService) MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.MarkRead(ctx, id, userID)
}

func (s *notificationService) Notify(ctx context.Context, userID uuid.UUID, notifType, title, message string, data map[string]interface{}) error {
	var dataJSON *string
	if len(data) > 0 {
		if b, err := json.Marshal(data); err == nil {
			s := string(b)
			dataJSON = &s
		}
	}

	notification := entities.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: message,
		Data:    dataJSON,
	}
	if err := s.repo.Create(ctx, &notification); err != nil {
		return err
	}

	s.broadcastWS(userID, notification)
	s.sendPush(ctx, userID, title, message, data)

	return nil
}

func (s *notificationService) broadcastWS(userID uuid.UUID, notification entities.Notification) {
	if s.wsService == nil {
		return
	}

	event := struct {
		Event string                   `json:"event"`
		Data  dto.NotificationResponse `json:"data"`
	}{
		Event: "NOTIFICATION",
		Data:  toResponse(notification),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	s.wsService.BroadcastToUsers([]string{userID.String()}, payload)
}

func (s *notificationService) sendPush(ctx context.Context, userID uuid.UUID, title, message string, data map[string]interface{}) {
	if s.push == nil {
		return
	}

	tokens, err := s.tokenRepo.FindTokensByUserID(ctx, userID)
	if err != nil || len(tokens) == 0 {
		return
	}

	strData := make(map[string]string, len(data))
	for k, v := range data {
		strData[k] = fmt.Sprintf("%v", v)
	}

	for _, token := range tokens {
		if err := s.push.Send(ctx, fcm.Message{Token: token, Title: title, Body: message, Data: strData}); err != nil {
			log.Printf("[notification] error enviando push a %s: %v", userID, err)

			var sendErr *fcm.SendError
			if errors.As(err, &sendErr) && sendErr.IsInvalidToken() {
				_ = s.tokenRepo.DeleteByToken(ctx, token)
			}
		}
	}
}

func toResponse(n entities.Notification) dto.NotificationResponse {
	return dto.NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Data:      n.Data,
		Read:      n.Read,
		CreatedAt: n.CreatedAt,
	}
}
