package grpc_server

import (
	"log"
	"net"

	"github.com/Caknoooo/go-gin-clean-starter/modules/device/grpc_handler"
	"github.com/Caknoooo/go-gin-clean-starter/modules/device/service"

	pos_handler "github.com/Caknoooo/go-gin-clean-starter/modules/position/grpc_handler"
	pos_service "github.com/Caknoooo/go-gin-clean-starter/modules/position/service"

	pbDevice "github.com/Caknoooo/go-gin-clean-starter/pkg/pb/device_proto"
	pbPosition "github.com/Caknoooo/go-gin-clean-starter/pkg/pb/position_proto"

	providerWS "github.com/Caknoooo/go-gin-clean-starter/providers/websocket"
	"gorm.io/gorm"
	"google.golang.org/grpc"
)

func StartGRPCServer(devService service.DeviceService, posService pos_service.PositionService, wsService providerWS.WebsocketService, db *gorm.DB, port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()

	// Inicializar Handlers
	devHandler := grpc_handler.NewDeviceGRPCHandler(devService)
	positionHandler := pos_handler.NewPositionGRPCHandler(posService, wsService, db)

	// Registrar Servicios
	pbDevice.RegisterDeviceServiceServer(grpcServer, devHandler)
	pbPosition.RegisterPositionServiceServer(grpcServer, positionHandler)

	log.Printf("Starting gRPC server on %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
