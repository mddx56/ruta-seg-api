package websocket

import (
	"log"
	"sync"
)

type Hub struct {
	// Registered clients.
	// Map userId -> map of clients. A user can have multiple connections (devices).
	Clients map[string]map[*Client]bool
	Admins  map[*Client]bool

	// Mutex for concurrent map access
	mu sync.RWMutex

	// Inbound messages from the clients.
	Broadcast chan *Message

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

type Message struct {
	TargetUsers []string // Specific UserIDs to broadcast to. If empty and IsGlobal=false, only sends to admins.
	IsGlobal    bool
	Payload     []byte
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]map[*Client]bool),
		Admins:     make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Clients[client.UserID] == nil {
				h.Clients[client.UserID] = make(map[*Client]bool)
			}
			h.Clients[client.UserID][client] = true
			
			if client.Role == "admin" {
				h.Admins[client] = true
			}
			h.mu.Unlock()
			log.Println("WS Client connected. UserID:", client.UserID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID][client]; ok {
				delete(h.Clients[client.UserID], client)
				if client.Role == "admin" {
					delete(h.Admins, client)
				}
				close(client.Send)
				if len(h.Clients[client.UserID]) == 0 {
					delete(h.Clients, client.UserID)
				}
				log.Println("WS Client disconnected. UserID:", client.UserID)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			if message.IsGlobal {
				// Broadcast to all connected clients
				for _, userClients := range h.Clients {
					for client := range userClients {
						select {
						case client.Send <- message.Payload:
						default:
							// Client queue full, meaning dead/stuck. Unregister.
							close(client.Send)
							delete(userClients, client)
						}
					}
				}
			} else {
				// Send only to specific target users AND all admins.
				sentTo := make(map[*Client]bool) // Evitar duplicados si admin está en target list.
				
				// Send to target users
				for _, uid := range message.TargetUsers {
					if userClients, ok := h.Clients[uid]; ok {
						for client := range userClients {
							if sentTo[client] { continue }
							sentTo[client] = true
							select {
							case client.Send <- message.Payload:
							default:
								close(client.Send)
								delete(userClients, client)
								if client.Role == "admin" {
									delete(h.Admins, client)
								}
							}
						}
					}
				}
				
				// Send to admins
				for admin := range h.Admins {
					if sentTo[admin] { continue }
					sentTo[admin] = true
					select {
					case admin.Send <- message.Payload:
					default:
						close(admin.Send)
						delete(h.Admins, admin)
						if admin.UserID != "" {
							delete(h.Clients[admin.UserID], admin)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}
