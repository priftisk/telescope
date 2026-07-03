package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"telescope/internal"
	"telescope/internal/container"
	router "telescope/internal/router"
	"time"

	"github.com/moby/moby/client"
)

type Server struct {
	dockerClient *client.Client
	routeTable   *router.RouteTable
	startTime    time.Time

	// Internal state
	httpServer  *http.Server
	proxyServer *http.Server
	wg          sync.WaitGroup
}

func NewServer() (*Server, error) {
	// Initialize logger
	// InitLogger()

	// Create Docker client
	apiClient, err := client.New(
		client.FromEnv,
		client.WithUserAgent("telescope/1.0.0"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Server{
		dockerClient: apiClient,
		routeTable:   &router.RouteTable{},
		startTime:    time.Now(),
	}, nil

}

func (s *Server) onStartup(ctx context.Context) error {
	slog.Info("Seeding route table")
	containers, err := s.dockerClient.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		log.Fatal("failed to list containers:", err)
	}
	for _, c := range containers.Items {
		info, err := s.dockerClient.ContainerInspect(ctx, c.ID, client.ContainerInspectOptions{})
		if err != nil {
			log.Printf("inspect failed for %s: %v", c.ID, err)
			continue
		}
		container, valid := container.ExtractContainerData(info.Container)
		if !valid {
			continue
		}
		s.routeTable.Register(container)
	}
	return nil
}

func (s *Server) serve(ctx context.Context) error {
	// --- Dashboard / API server ---
	dashboardMux := http.NewServeMux()
	// API
	dashboardMux.HandleFunc("GET /routes", s.RoutesHandler)
	// Dashboard
	dashboardMux.HandleFunc("GET /dashboard", s.DashboardHandler)
	dashboardMux.HandleFunc("/dashboard/{resource}", s.DashboardResourceHandler)
	static_dir, _ := internal.GetStaticDir()
	dashboardMux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(static_dir))))

	dashboardServer := &http.Server{
		Addr:    ":8900",
		Handler: dashboardMux,
	}

	// --- Proxy server (catch-all) ---
	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("/", s.ProxyHandler)

	proxyServer := &http.Server{
		Addr:    ":8901",
		Handler: proxyMux,
	}

	s.httpServer = dashboardServer
	s.proxyServer = proxyServer

	// Handle shutdown for both servers in background
	go func() {
		<-ctx.Done()
		slog.Info("Shutting down HTTP servers...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			if err := dashboardServer.Shutdown(shutdownCtx); err != nil {
				slog.Error("Dashboard server shutdown error", "error", err)
			}
		}()

		go func() {
			defer wg.Done()
			if err := proxyServer.Shutdown(shutdownCtx); err != nil {
				slog.Error("Proxy server shutdown error", "error", err)
			}
		}()

		wg.Wait()
	}()

	// slog.Info("Telescope dashboard listening on :8900")
	slog.Info("Dashboard available at http://localhost:8900/dashboard")
	slog.Info("Telescope proxy listening on :8901")

	// Run both servers concurrently; return on first fatal error
	errCh := make(chan error, 2)

	go func() {
		if err := dashboardServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("dashboard server error: %w", err)
			return
		}
		errCh <- nil
	}()

	go func() {
		if err := proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("proxy server error: %w", err)
			return
		}
		errCh <- nil
	}()

	// Wait for both to finish (either cleanly on shutdown, or with an error)
	var firstErr error
	for range 2 {
		if err := <-errCh; err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *Server) Run() error {
	// Setup signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Ensure cleanup
	defer s.dockerClient.Close()

	slog.Info("🔭 Telescope starting...")

	// 1. Seed the route table — blocking, must finish first
	if err := s.onStartup(ctx); err != nil {
		return fmt.Errorf("failed to seed routes: %w", err)
	}

	// 2. Launch event watcher in background
	s.wg.Go(func() {
		s.watchEvents(ctx)
	})

	// 3. Start HTTP and UI servers — blocking, keeps main alive
	if err := s.serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	// 4. Wait for event watcher to finish after shutdown
	s.wg.Wait()

	slog.Info("Telescope shutdown complete")
	return nil
}
