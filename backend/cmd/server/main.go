package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"poker-grpc/backend/internal/server"
	pokerv1 "poker-grpc/backend/poker-grpc/gen/poker/v1"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	grpcServer := grpc.NewServer()
	pokerv1.RegisterPokerServiceServer(grpcServer, server.NewPokerServer())

	log.Printf("gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
