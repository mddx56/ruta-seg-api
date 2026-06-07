package grpc_handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/modules/position/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position/service"
	pb "github.com/Caknoooo/go-gin-clean-starter/pkg/pb/position_proto"
	providerWS "github.com/Caknoooo/go-gin-clean-starter/providers/websocket"
	"gorm.io/gorm"
)

type PositionGRPCHandler struct {
	pb.UnimplementedPositionServiceServer
	positionService service.PositionService
	wsService       providerWS.WebsocketService
	db              *gorm.DB
}

func NewPositionGRPCHandler(s service.PositionService, wsSvc providerWS.WebsocketService, db *gorm.DB) *PositionGRPCHandler {
	return &PositionGRPCHandler{
		positionService: s,
		wsService:       wsSvc,
		db:              db,
	}
}

type broadcastAttrs struct {
	battery    *float64
	ignition   *bool
	satellites *int
}

func extractBroadcastAttributes(raw *string) broadcastAttrs {
	if raw == nil {
		return broadcastAttrs{}
	}
	attrs := struct {
		Battery    *float64 `json:"battery"`
		Ignition   *bool    `json:"ignition"`
		Satellites *int     `json:"satellites"`
	}{}
	_ = json.Unmarshal([]byte(*raw), &attrs)
	return broadcastAttrs{
		battery:    attrs.Battery,
		ignition:   attrs.Ignition,
		satellites: attrs.Satellites,
	}
}

func (h *PositionGRPCHandler) SavePosition(ctx context.Context, req *pb.SavePositionRequest) (*pb.SavePositionResponse, error) {
	// Parse attributes from string to *string if not empty
	var attributes *string
	if req.Attributes != "" {
		attrs := req.Attributes
		attributes = &attrs
	}

	deviceTime := time.Unix(req.DeviceTime, 0)

	createReq := dto.PositionCreateRequest{
		Imei:       req.Imei,
		DeviceTime: deviceTime,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Speed:      int(req.Speed),
		Course:     int(req.Course),
		Attributes: attributes,
	}

	result, err := h.positionService.Create(ctx, createReq)
	if err != nil {
		return &pb.SavePositionResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Broadcast WebSocket
	if h.wsService != nil {
		var userIDs []string

		h.db.WithContext(ctx).
			Table("device_installations").
			Select("vehicles.user_id").
			Joins("JOIN vehicles ON vehicles.id = device_installations.vehicle_id").
			Where("device_installations.imei = ? AND device_installations.removed_at IS NULL AND device_installations.status = ?", req.Imei, true).
			Pluck("vehicles.user_id", &userIDs)

		parsedAttrs := extractBroadcastAttributes(attributes)
		go h.wsService.BroadcastPosition(userIDs, providerWS.DevicePositionData{
			IMEI:       result.Imei,
			Latitude:   result.Latitude,
			Longitude:  result.Longitude,
			Speed:      result.Speed,
			Course:     result.Course,
			DeviceTime: result.DeviceTime,
			ServerTime: result.ServerTime,
			Battery:    parsedAttrs.battery,
			Ignition:   parsedAttrs.ignition,
			Satellites: parsedAttrs.satellites,
		})
	}

	return &pb.SavePositionResponse{
		Success: true,
		Message: "Posición guardada exitosamente",
	}, nil
}
