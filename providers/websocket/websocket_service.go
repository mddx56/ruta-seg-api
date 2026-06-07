package websocket

type WebsocketService interface {
	RunHub()
	GetHub() *Hub
	BroadcastToUsers(userIDs []string, payload []byte)
	BroadcastToAll(payload []byte)
	BroadcastPosition(userIDs []string, data DevicePositionData)
}

type websocketService struct {
	hub *Hub
}

func NewWebsocketService(hub *Hub) WebsocketService {
	return &websocketService{
		hub: hub,
	}
}

func (s *websocketService) RunHub() {
	s.hub.Run()
}

func (s *websocketService) GetHub() *Hub {
	return s.hub
}

func (s *websocketService) BroadcastToUsers(userIDs []string, payload []byte) {
	msg := &Message{
		TargetUsers: userIDs,
		IsGlobal:    false,
		Payload:     payload,
	}
	s.hub.Broadcast <- msg
}

func (s *websocketService) BroadcastToAll(payload []byte) {
	msg := &Message{
		IsGlobal: true,
		Payload:  payload,
	}
	s.hub.Broadcast <- msg
}
