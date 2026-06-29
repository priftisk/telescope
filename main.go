package main

import (
	"context"
	"log"
	"sync"

	"github.com/moby/moby/client"
)

func main() {
	parentContext := context.Background()
	ctx, cancel := context.WithCancel(parentContext)
	defer cancel()

	apiClient, err := client.New(
		client.FromEnv,
		client.WithUserAgent("telescope/1.0.0"),
	)
	if err != nil {
		log.Fatal("failed to create docker client:", err)
	}
	defer apiClient.Close()

	var rt *RouteTable = &RouteTable{}

	// 1. Seed the route table — blocking, must finish first
	onStartup(ctx, apiClient, rt)

	// 2. Launch event watcher in background
	var wg sync.WaitGroup
	wg.Go(func() {
		watchEvents(ctx, apiClient, rt)
	})

	// 3. Start HTTP proxy — blocking, keeps main alive
	// when ctx is cancelled, runProxy calls server.Shutdown
	runProxy(ctx, rt)

	// 4. Wait for event watcher to finish after shutdown
	wg.Wait()
}
