package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/moby/moby/client"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	InitLogger()

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
