package main

import (
	"context"
	"log"

	"github.com/moby/moby/client"
)

func main() {
	ctx := context.Background()

	apiClient, err := client.New(
		client.FromEnv,
		client.WithUserAgent("telescope/1.0.0"),
	)
	if err != nil {
		log.Fatal("failed to create docker client:", err)
	}
	defer apiClient.Close()

	var rt *RouteTable = &RouteTable{}

	onStartup(ctx, apiClient, rt)

	watchEvents(ctx, apiClient, rt)
}
