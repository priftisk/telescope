package main

import (
	"log"
	"telescope/internal/config"
	"telescope/internal/server"
)

func main() {
	server, err := server.NewServer(
		config.WithProxyAddress("0.0.0.0:8999"),
		config.WithDashboardAddress("0.0.0.0:8925"),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

}
