package main

import (
	"log"
	"telescope/internal/server"
)

func main() {
	server, err := server.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

}
