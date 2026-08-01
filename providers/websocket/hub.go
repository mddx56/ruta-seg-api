package websocket

import (
	"log"
	"sync"

	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
)

// maxConnectionsPerUser evita que una app con bug/loop de reconexión agote memoria.
const maxConnectionsPerUser = 10

type Hub struct {
	// Clientes registrados: userID -> clientes (un usuario puede tener varias conexiones).
	clients map[string]map[*Client]bool
	admins  map[*Client]bool

	// Suscriptores públicos sin login, agrupados por topic (ej. "route:<routeID>").
	topics map[string]map[*Client]bool

	// Protege los mapas de acceso concurrente.
	mu sync.RWMutex

	// Mensajes a difundir a clientes (dirigidos o globales).
	Broadcast chan *Message

	// Mensajes dirigidos a un topic público.
	TopicBroadcast chan *TopicMessage

	// Solicitudes de registro de clientes.
	Register chan *Client

	// Solicitudes de baja de clientes.
	Unregister chan *Client
}

type Message struct {
	TargetUsers []string // IDs de usuario destino; si está vacío y no es global, solo llega a admins.
	IsGlobal    bool
	Payload     []byte
	VehicleID   string // Si no está vacío, filtra a qué admins llega (ver broadcastTargeted).
}

type TopicMessage struct {
	Topic   string
	Payload []byte
}

// Stats es una foto de cuántos clientes hay conectados, para debug/monitoreo.
type Stats struct {
	Users  int // Usuarios distintos con al menos una conexión (sin contar topics).
	Admins int
	Topics int
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:      make(chan *Message),
		TopicBroadcast: make(chan *TopicMessage),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		clients:        make(map[string]map[*Client]bool),
		admins:         make(map[*Client]bool),
		topics:         make(map[string]map[*Client]bool),
	}
}

// Stats devuelve un snapshot de conexiones activas, protegido por el mutex del hub.
func (h *Hub) Stats() Stats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return Stats{
		Users:  len(h.clients),
		Admins: len(h.admins),
		Topics: len(h.topics),
	}
}

// Aísla cada caso: un panic no debe tumbar el hub ni dejar el mutex trabado.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.handleRegister(client)
		case client := <-h.Unregister:
			h.handleUnregister(client)
		case topicMsg := <-h.TopicBroadcast:
			h.handleTopicBroadcast(topicMsg)
		case message := <-h.Broadcast:
			h.handleBroadcast(message)
		}
	}
}

// recoverPanic loguea y absorbe un panic para que el hub siga vivo.
func (h *Hub) recoverPanic(where string) {
	if r := recover(); r != nil {
		log.Printf("hub: panic recuperado en %s: %v", where, r)
	}
}

func (h *Hub) handleRegister(client *Client) {
	defer h.recoverPanic("register")
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.Topic != "" {
		if h.topics[client.Topic] == nil {
			h.topics[client.Topic] = make(map[*Client]bool)
		}
		h.topics[client.Topic][client] = true
		log.Println("WS Client connected. Topic:", client.Topic)
		return
	}

	if len(h.clients[client.UserID]) >= maxConnectionsPerUser {
		// No tocar client.Conn: WritePump ya corre en paralelo y es el único que debe escribirle.
		log.Printf("WS: usuario %s alcanzó el límite de %d conexiones, se rechaza", client.UserID, maxConnectionsPerUser)
		close(client.Send)
		return
	}

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true

	if client.Role == constants.ENUM_ROLE_ADMIN {
		h.admins[client] = true
	}
	log.Println("WS Client connected. UserID:", client.UserID)
}

func (h *Hub) handleUnregister(client *Client) {
	defer h.recoverPanic("unregister")
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.Topic != "" {
		if _, ok := h.topics[client.Topic][client]; ok {
			delete(h.topics[client.Topic], client)
			close(client.Send)
			if len(h.topics[client.Topic]) == 0 {
				delete(h.topics, client.Topic)
			}
			log.Println("WS Client disconnected. Topic:", client.Topic)
		}
	} else if _, ok := h.clients[client.UserID][client]; ok {
		delete(h.clients[client.UserID], client)
		if client.Role == constants.ENUM_ROLE_ADMIN {
			delete(h.admins, client)
		}
		close(client.Send)
		if len(h.clients[client.UserID]) == 0 {
			delete(h.clients, client.UserID)
		}
		log.Println("WS Client disconnected. UserID:", client.UserID)
	}
}

func (h *Hub) handleTopicBroadcast(topicMsg *TopicMessage) {
	defer h.recoverPanic("topic-broadcast")
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.topics[topicMsg.Topic] {
		select {
		case client.Send <- topicMsg.Payload:
		default:
			// Buffer lleno: cliente colgado, se descarta.
			close(client.Send)
			delete(h.topics[topicMsg.Topic], client)
		}
	}
	if len(h.topics[topicMsg.Topic]) == 0 {
		delete(h.topics, topicMsg.Topic)
	}
}

func (h *Hub) handleBroadcast(message *Message) {
	defer h.recoverPanic("broadcast")
	h.mu.Lock()
	defer h.mu.Unlock()

	if message.IsGlobal {
		h.broadcastGlobal(message.Payload)
	} else {
		h.broadcastTargeted(message)
	}
}

// broadcastGlobal difunde a todos los clientes conectados.
func (h *Hub) broadcastGlobal(payload []byte) {
	for uid, userClients := range h.clients {
		for client := range userClients {
			select {
			case client.Send <- payload:
			default:
				// Buffer lleno: cliente colgado, se descarta.
				close(client.Send)
				delete(userClients, client)
				if client.Role == constants.ENUM_ROLE_ADMIN {
					delete(h.admins, client)
				}
			}
		}
		if len(userClients) == 0 {
			delete(h.clients, uid)
		}
	}
}

// broadcastTargeted envía a los usuarios destino y a todos los admins.
func (h *Hub) broadcastTargeted(message *Message) {
	sentTo := make(map[*Client]bool) // Evita duplicados si el admin también es destino.

	for _, uid := range message.TargetUsers {
		userClients, ok := h.clients[uid]
		if !ok {
			continue
		}
		for client := range userClients {
			if sentTo[client] {
				continue
			}
			sentTo[client] = true
			select {
			case client.Send <- message.Payload:
			default:
				close(client.Send)
				delete(userClients, client)
				if client.Role == constants.ENUM_ROLE_ADMIN {
					delete(h.admins, client)
				}
			}
		}
		if len(userClients) == 0 {
			delete(h.clients, uid)
		}
	}

	for admin := range h.admins {
		if sentTo[admin] {
			continue
		}
		if message.VehicleID != "" && admin.AllowedVehicleIDs != nil && !admin.AllowedVehicleIDs[message.VehicleID] {
			continue // admin filtró su flota y este vehículo no está en su lista
		}
		sentTo[admin] = true
		select {
		case admin.Send <- message.Payload:
		default:
			close(admin.Send)
			delete(h.admins, admin)
			if admin.UserID != "" {
				delete(h.clients[admin.UserID], admin)
				if len(h.clients[admin.UserID]) == 0 {
					delete(h.clients, admin.UserID)
				}
			}
		}
	}
}
