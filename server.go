package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/moby/moby/client"
)

type Server struct {
	dockerClient *client.Client
	routeTable   *RouteTable
	startTime    time.Time

	// Internal state
	httpServer *http.Server
	uiServer   *http.Server
	wg         sync.WaitGroup
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
		routeTable:   &RouteTable{},
		startTime:    time.Now(),
	}, nil

}

func StripPort(host string) string {
	hostOnly, _, err := net.SplitHostPort(host)
	if err != nil {
		// No port present.
		return host
	}
	return hostOnly
}

func GetHostAndPath(r *http.Request) (string, string) {
	host := StripPort(r.Host)

	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		return host, "/"
	}

	if i := strings.IndexByte(path, '/'); i != -1 {
		path = path[:i]
	}

	return host, path
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
		container, valid := ExtractContainerData(info.Container)
		if !valid {
			continue
		}
		s.routeTable.Register(container)
	}
	return nil
}

func (s *Server) serve(ctx context.Context) error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("GET /routes", s.RoutesHandler)

	// Dashboard UI
	mux.HandleFunc("GET /dashboard", s.DashboardHandler)
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./static"))))

	// Proxy - catch-all must be registered last
	mux.HandleFunc("/", s.ProxyHandler)

	s.httpServer = &http.Server{
		Addr:    ":8901",
		Handler: mux,
	}

	// Handle shutdown in background
	go func() {
		<-ctx.Done()
		slog.Info("Shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}
	}()

	slog.Info("Telescope listening on :8901")
	slog.Info("Dashboard available at http://localhost:8901/dashboard")

	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
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

	// 3. Start HTTP server — blocking, keeps main alive
	if err := s.serve(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	// 4. Wait for event watcher to finish after shutdown
	s.wg.Wait()

	slog.Info("Telescope shutdown complete")
	return nil
}
