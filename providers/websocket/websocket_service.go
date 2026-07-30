package websocket

type WebsocketService interface {
	RunHub()
	GetHub() *Hub
	BroadcastToUsers(userIDs []string, payload []byte)
	BroadcastToAll(payload []byte)
	BroadcastPosition(userIDs []string, data DevicePositionData)
	// BroadcastToTopic envía payload a los clientes públicos suscritos a topic
	// (usado por el canal público de micros en ruta, sin pasar por Redis: solo llega
	// a los clientes conectados a esta misma instancia — ver RouteEventPublisher para
	// el fan-out entre instancias).
	BroadcastToTopic(topic string, payload []byte)
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

func (s *websocketService) BroadcastToTopic(topic string, payload []byte) {
	s.hub.TopicBroadcast <- &TopicMessage{
		Topic:   topic,
		Payload: payload,
	}
}
