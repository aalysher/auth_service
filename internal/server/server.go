package server

import (
	"fmt"
	"log"
	"net"

	"github.com/aalysher/auth_service/config"
	"github.com/aalysher/auth_service/handler"
	pb "github.com/aalysher/auth_service/proto"

	"google.golang.org/grpc"
)

func RunServer() {
	address := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", address, err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &handler.AuthHandler{})

	log.Printf("Starting gRPC server on %s...", address)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
