package websocket

import (
	"encoding/json"
	"log"
	"time"
)

// DevicePositionEvent is the payload sent over the WebSocket to all clients
// when a device sends a new GPS position.
type DevicePositionEvent struct {
	Event string             `json:"event"`
	Data  DevicePositionData `json:"data"`
}

type DevicePositionData struct {
	IMEI       string    `json:"imei"`
	VehicleID  string    `json:"vehicle_id,omitempty"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Speed      int       `json:"speed"`
	Course     int       `json:"course"`
	DeviceTime time.Time `json:"device_time"`
	ServerTime time.Time `json:"server_time"`
	Battery    *float64  `json:"battery,omitempty"`
	Ignition   *bool     `json:"ignition,omitempty"`
	Satellites *int      `json:"satellites,omitempty"`
	Category   string    `json:"category"`
}

// broadcastThrottleWindow evita saturar a los admins con un DEVICE_UPDATED por cada ping crudo del GPS.
const broadcastThrottleWindow = 3 * time.Second

// BroadcastPosition sends a DEVICE_UPDATED event to specific users and all admins.
// Category is determined by speed/ignition logic here.
func (s *websocketService) BroadcastPosition(userIDs []string, data DevicePositionData) {
	if !s.allowBroadcast(data.IMEI) {
		return
	}

	data.Category = resolveCategory(data.Speed, data.Ignition)

	event := DevicePositionEvent{
		Event: "DEVICE_UPDATED",
		Data:  data,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("ws: error marshaling position event:", err)
		return
	}

	s.hub.Broadcast <- &Message{
		TargetUsers: userIDs,
		IsGlobal:    false,
		Payload:     payload,
		VehicleID:   data.VehicleID,
	}
}

// allowBroadcast limita a un DEVICE_UPDATED por IMEI cada broadcastThrottleWindow.
func (s *websocketService) allowBroadcast(imei string) bool {
	now := time.Now()

	s.broadcastMu.Lock()
	defer s.broadcastMu.Unlock()

	if last, ok := s.lastBroadcastByID[imei]; ok && now.Sub(last) < broadcastThrottleWindow {
		return false
	}
	s.lastBroadcastByID[imei] = now
	return true
}

func resolveCategory(speed int, ignition *bool) string {
	const minMovingSpeed = 5
	if speed > minMovingSpeed {
		return "live"
	}
	if ignition != nil && *ignition {
		return "idling"
	}
	return "parked"
}
