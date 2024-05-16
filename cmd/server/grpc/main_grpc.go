package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/elina-chertova/metrics-alerting.git/api/proto"
	"github.com/elina-chertova/metrics-alerting.git/internal/config"
	handler "github.com/elina-chertova/metrics-alerting.git/internal/handlers/grpc"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/db"
	"github.com/elina-chertova/metrics-alerting.git/internal/storage/filememory"

	"google.golang.org/grpc"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version:%s\n", buildVersion)
	fmt.Printf("Build date:%s\n", buildDate)
	fmt.Printf("Build commit:%s\n", buildCommit)

	serverConfig := config.NewServer()
	h := buildStorageGRPC(serverConfig)

	lis, err := net.Listen("tcp", ":"+serverConfig.GRPCPort)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMetricsServiceServer(
		grpcServer,
		&handler.Server{
			Handler:   h,
			SecretKey: serverConfig.SecretKey,
			CryptoKey: serverConfig.CryptoKey,
		},
	)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-quit

		fmt.Println("Shutting down server...")

		grpcServer.GracefulStop()

		fmt.Println("Server exiting")
	}()

	log.Printf("gRPC server listening on %d", serverConfig.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func buildStorageGRPC(config *config.Server) *handler.Handler {
	if config.DatabaseDSN != "" {
		connection := db.Connect(config.DatabaseDSN)
		return handler.NewHandler(connection)
	} else {
		s := filememory.NewMemStorage(true, config)
		return handler.NewHandler(s)
	}
}
