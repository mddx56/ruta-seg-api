package grpc_handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	lapservice "github.com/Caknoooo/go-gin-clean-starter/modules/lap/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/position/service"
	pb "github.com/Caknoooo/go-gin-clean-starter/pkg/pb/position_proto"
	providerWS "github.com/Caknoooo/go-gin-clean-starter/providers/websocket"
)

type PositionGRPCHandler struct {
	pb.UnimplementedPositionServiceServer
	positionService service.PositionService
	wsService       providerWS.WebsocketService
	lapService      lapservice.LapService
	ownerResolver   service.DeviceOwnerResolver
}

func NewPositionGRPCHandler(s service.PositionService, wsSvc providerWS.WebsocketService, lapSvc lapservice.LapService, resolver service.DeviceOwnerResolver) *PositionGRPCHandler {
	return &PositionGRPCHandler{
		positionService: s,
		wsService:       wsSvc,
		lapService:      lapSvc,
		ownerResolver:   resolver,
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

	h.broadcastAndEvaluate(req.Imei, attributes, result)

	return &pb.SavePositionResponse{
		Success: true,
		Message: "Posición guardada exitosamente",
	}, nil
}

// broadcastAndEvaluate resuelve el dueño (con caché) y notifica en background, sin demorar la respuesta gRPC.
func (h *PositionGRPCHandler) broadcastAndEvaluate(imei string, attributes *string, result dto.PositionResponse) {
	if h.wsService == nil && h.lapService == nil {
		return
	}

	go func() {
		ctx := context.Background()
		owner, found := h.ownerResolver.Resolve(ctx, imei)

		if h.wsService != nil {
			var userIDs []string
			var vehicleID string
			if found {
				userIDs = []string{owner.UserID}
				vehicleID = owner.VehicleID.String()
			}
			parsedAttrs := extractBroadcastAttributes(attributes)
			h.wsService.BroadcastPosition(userIDs, providerWS.DevicePositionData{
				IMEI:       result.Imei,
				VehicleID:  vehicleID,
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

		if found && h.lapService != nil {
			if err := h.lapService.EvaluatePosition(ctx, owner.VehicleID, result.Latitude, result.Longitude, result.Speed, result.DeviceTime); err != nil {
				log.Printf("[lap-engine] error evaluando posición del vehículo %s: %v", owner.VehicleID, err)
			}
		}
	}()
}
