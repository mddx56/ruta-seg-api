package service

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/Caknoooo/go-gin-clean-starter/pkg/pb/device_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PeerChecker valida que un IMEI no esté registrado en el VPS par.
// Si PEER_GRPC_URL no está configurado, todas las consultas retornan false (deshabilitado).
type PeerChecker interface {
	IMEIExistsOnPeer(ctx context.Context, imei string) (bool, error)
	IMEIsExistOnPeer(ctx context.Context, imeis []string) (map[string]bool, error)
}

type peerChecker struct {
	client pb.DeviceServiceClient
	conn   *grpc.ClientConn
}

// NewPeerChecker crea el cliente gRPC hacia el VPS par definido en PEER_GRPC_URL.
// Retorna un checker deshabilitado (no-op) si la variable no está configurada.
func NewPeerChecker() PeerChecker {
	peerURL := os.Getenv("PEER_GRPC_URL")
	if peerURL == "" {
		return &nopPeerChecker{}
	}

	conn, err := grpc.NewClient(peerURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[PeerChecker] No se pudo conectar al VPS par (%s): %v — validación cruzada deshabilitada", peerURL, err)
		return &nopPeerChecker{}
	}

	log.Printf("[PeerChecker] Validación cruzada activa → %s", peerURL)
	return &peerChecker{
		client: pb.NewDeviceServiceClient(conn),
		conn:   conn,
	}
}

func (p *peerChecker) IMEIExistsOnPeer(ctx context.Context, imei string) (bool, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := p.client.CheckIMEIExists(timeoutCtx, &pb.CheckIMEIRequest{Imei: imei})
	if err != nil {
		return false, err
	}
	return resp.Exists, nil
}

func (p *peerChecker) IMEIsExistOnPeer(ctx context.Context, imeis []string) (map[string]bool, error) {
	result := make(map[string]bool)
	if len(imeis) == 0 {
		return result, nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := p.client.BatchCheckIMEIs(timeoutCtx, &pb.BatchCheckIMEIsRequest{Imeis: imeis})
	if err != nil {
		return result, err
	}
	for _, imei := range resp.Found {
		result[imei] = true
	}
	return result, nil
}

// nopPeerChecker se usa cuando PEER_GRPC_URL no está configurado.
type nopPeerChecker struct{}

func (n *nopPeerChecker) IMEIExistsOnPeer(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (n *nopPeerChecker) IMEIsExistOnPeer(_ context.Context, _ []string) (map[string]bool, error) {
	return make(map[string]bool), nil
}
