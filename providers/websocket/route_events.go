package websocket

import (
	"context"
	"encoding/json"
	"log"

	redisProvider "github.com/Caknoooo/go-gin-clean-starter/providers/redis"
)

// RouteEventsChannel es el canal Redis Pub/Sub usado para propagar eventos de micros
// (posición, vuelta completada) entre todas las instancias del servidor, de forma que
// el canal público de WebSocket (sin login, para TV y app pública) funcione igual sin
// importar a qué instancia esté conectado cada cliente.
const RouteEventsChannel = "micros:route-events"

// TopicAll agrupa los eventos de todas las rutas, útil para un mapa público con varias líneas.
const TopicAll = "route:all"

// RouteEvent es el envoltorio que viaja por Redis; Topic decide a qué suscriptores
// locales llega en cada instancia.
type RouteEvent struct {
	Topic string      `json:"topic"`
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type routeClientPayload struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// RouteEventPublisher publica eventos de una ruta (posición en vivo, vuelta completada,
// etc.) para que los reciban todos los suscriptores del canal público, en cualquier
// instancia del servidor.
type RouteEventPublisher interface {
	Publish(ctx context.Context, topic, event string, data interface{}) error
}

type routeEventPublisher struct {
	redis redisProvider.RedisService
}

func NewRouteEventPublisher(redisSvc redisProvider.RedisService) RouteEventPublisher {
	return &routeEventPublisher{redis: redisSvc}
}

func (p *routeEventPublisher) Publish(ctx context.Context, topic, event string, data interface{}) error {
	payload, err := json.Marshal(RouteEvent{Topic: topic, Event: event, Data: data})
	if err != nil {
		return err
	}
	return p.redis.Publish(ctx, RouteEventsChannel, payload)
}

// StartRouteEventSubscriber corre en background (una sola vez por instancia) suscrito
// al canal Redis y reenvía cada evento al Hub local bajo su topic específico y bajo
// TopicAll, para los clientes públicos conectados a esta instancia.
func StartRouteEventSubscriber(ctx context.Context, redisSvc redisProvider.RedisService, hub *Hub) {
	pubsub := redisSvc.Client().Subscribe(ctx, RouteEventsChannel)

	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()
		for msg := range ch {
			var event RouteEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("[route-events] payload inválido: %v", err)
				continue
			}

			clientPayload, err := json.Marshal(routeClientPayload{Event: event.Event, Data: event.Data})
			if err != nil {
				log.Printf("[route-events] error serializando payload para clientes: %v", err)
				continue
			}

			hub.TopicBroadcast <- &TopicMessage{Topic: event.Topic, Payload: clientPayload}
			if event.Topic != TopicAll {
				hub.TopicBroadcast <- &TopicMessage{Topic: TopicAll, Payload: clientPayload}
			}
		}
	}()
}
