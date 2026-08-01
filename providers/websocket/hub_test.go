package websocket

import (
	"testing"

	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
)

func isClosed(ch chan []byte) bool {
	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}

func TestHandleRegisterAndUnregister(t *testing.T) {
	h := NewHub()
	client := &Client{UserID: "u1", Role: constants.ENUM_ROLE_ADMIN, Send: make(chan []byte, 4)}

	h.handleRegister(client)
	if !h.clients["u1"][client] {
		t.Fatal("el cliente debería estar en clients")
	}
	if !h.admins[client] {
		t.Fatal("el admin debería estar en admins")
	}

	h.handleUnregister(client)
	if _, ok := h.clients["u1"]; ok {
		t.Fatal("la clave del usuario debería borrarse al quedar vacía")
	}
	if h.admins[client] {
		t.Fatal("el admin debería salir de admins al desconectarse")
	}
	if !isClosed(client.Send) {
		t.Fatal("Send debería quedar cerrado tras el unregister")
	}
}

func TestHandleRegisterConnectionLimit(t *testing.T) {
	h := NewHub()
	const uid = "u2"

	for range maxConnectionsPerUser {
		h.handleRegister(&Client{UserID: uid, Send: make(chan []byte, 1)})
	}
	if got := len(h.clients[uid]); got != maxConnectionsPerUser {
		t.Fatalf("esperaba %d conexiones registradas, hay %d", maxConnectionsPerUser, got)
	}

	extra := &Client{UserID: uid, Send: make(chan []byte, 1)}
	h.handleRegister(extra)

	if got := len(h.clients[uid]); got != maxConnectionsPerUser {
		t.Fatalf("la conexión %d no debería haberse aceptado, hay %d registradas", maxConnectionsPerUser+1, got)
	}
	if !isClosed(extra.Send) {
		t.Fatal("Send del cliente rechazado debería quedar cerrado")
	}
}

func TestBroadcastGlobalFullBufferRemovesAdmin(t *testing.T) {
	h := NewHub()
	admin := &Client{UserID: "a1", Role: constants.ENUM_ROLE_ADMIN, Send: make(chan []byte)} // sin buffer: siempre "lleno"
	h.handleRegister(admin)

	h.broadcastGlobal([]byte("hola"))

	if h.admins[admin] {
		t.Fatal("el admin con buffer lleno debe salir de admins (bug original: send on closed channel)")
	}
	if _, ok := h.clients["a1"]; ok {
		t.Fatal("la clave del usuario debería borrarse al quedar vacía")
	}
	if !isClosed(admin.Send) {
		t.Fatal("Send debería quedar cerrado")
	}
}

func TestBroadcastTargetedFullBufferRemovesAdmin(t *testing.T) {
	h := NewHub()
	admin := &Client{UserID: "a2", Role: constants.ENUM_ROLE_ADMIN, Send: make(chan []byte)}
	h.handleRegister(admin)

	h.broadcastTargeted(&Message{TargetUsers: []string{}, Payload: []byte("hola")})

	if h.admins[admin] {
		t.Fatal("el admin con buffer lleno debe salir de admins")
	}
	if _, ok := h.clients["a2"]; ok {
		t.Fatal("la clave del usuario debería borrarse al quedar vacía")
	}
}

func TestTopicBroadcastFullBufferCleansEmptyTopic(t *testing.T) {
	h := NewHub()
	client := &Client{Topic: "route:1", Send: make(chan []byte)}
	h.handleRegister(client)

	h.handleTopicBroadcast(&TopicMessage{Topic: "route:1", Payload: []byte("pos")})

	if _, ok := h.topics["route:1"]; ok {
		t.Fatal("el topic debería borrarse al quedar sin suscriptores")
	}
}

func TestBroadcastTargetedFiltersAdminsByVehicleID(t *testing.T) {
	h := NewHub()
	filtered := &Client{UserID: "af", Role: constants.ENUM_ROLE_ADMIN, AllowedVehicleIDs: map[string]bool{"v1": true}, Send: make(chan []byte, 4)}
	unfiltered := &Client{UserID: "au", Role: constants.ENUM_ROLE_ADMIN, Send: make(chan []byte, 4)}
	h.handleRegister(filtered)
	h.handleRegister(unfiltered)

	h.broadcastTargeted(&Message{Payload: []byte("v1-event"), VehicleID: "v1"})
	if len(filtered.Send) != 1 {
		t.Fatal("el admin filtrado debería recibir eventos de un vehículo que sí tiene permitido")
	}
	if len(unfiltered.Send) != 1 {
		t.Fatal("el admin sin filtro debería recibir todo")
	}

	h.broadcastTargeted(&Message{Payload: []byte("v2-event"), VehicleID: "v2"})
	if len(filtered.Send) != 1 {
		t.Fatal("el admin filtrado NO debería recibir eventos de un vehículo que no tiene permitido")
	}
	if len(unfiltered.Send) != 2 {
		t.Fatal("el admin sin filtro debería seguir recibiendo todo")
	}

	h.broadcastTargeted(&Message{Payload: []byte("notif"), VehicleID: ""})
	if len(filtered.Send) != 2 {
		t.Fatal("un mensaje sin VehicleID (ej. notificación) debe llegar a todos los admins igual")
	}
}

func TestStats(t *testing.T) {
	h := NewHub()
	h.handleRegister(&Client{UserID: "u1", Role: constants.ENUM_ROLE_ADMIN, Send: make(chan []byte, 1)})
	h.handleRegister(&Client{UserID: "u2", Send: make(chan []byte, 1)})
	h.handleRegister(&Client{Topic: "route:1", Send: make(chan []byte, 1)})

	stats := h.Stats()
	if stats.Users != 2 {
		t.Fatalf("esperaba 2 usuarios, hay %d", stats.Users)
	}
	if stats.Admins != 1 {
		t.Fatalf("esperaba 1 admin, hay %d", stats.Admins)
	}
	if stats.Topics != 1 {
		t.Fatalf("esperaba 1 topic, hay %d", stats.Topics)
	}
}
