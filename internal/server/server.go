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
	"telescope/internal/config"
	"telescope/internal/container"
	"telescope/internal/dashboard"
	"telescope/internal/proxy"
	"telescope/internal/roundtripper"
	router "telescope/internal/router"
	"time"

	"github.com/moby/moby/client"
)

// Handles the lifecycle of the app.
type Server struct {
	dockerClient *client.Client
	routeTable   *router.RouteTable
	startTime    time.Time

	// Internal state
	dashboard *dashboard.DashboardServer
	proxy     *proxy.ProxyServer
	trips     *roundtripper.Trips
	wg        sync.WaitGroup
}

func NewServer(opts ...config.Opt) (*Server, error) {
	apiClient, err := client.New(
		client.FromEnv,
		client.WithUserAgent("telescope/1.0.0"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	routeTable := router.NewRouteTable()
	trips := roundtripper.NewTripsRecorder()
	startTime := time.Now()

	dashboardSrv, err := dashboard.NewDashboardServer(routeTable, trips, startTime, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create dashboard server: %w", err)
	}
	proxySrv := proxy.NewProxyServer(routeTable, trips, opts...)

	return &Server{
		dockerClient: apiClient,
		routeTable:   routeTable,
		startTime:    startTime,
		dashboard:    dashboardSrv,
		proxy:        proxySrv,
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
	// Handle shutdown for both servers in background
	go func() {
		<-ctx.Done()
		slog.Info("Shutting down HTTP servers...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// var wg sync.WaitGroup
		s.wg.Add(2)

		go func() {
			defer s.wg.Done()
			if err := s.dashboard.Shutdown(shutdownCtx); err != nil {
				slog.Error("Dashboard server shutdown error", "error", err)
			}
		}()

		go func() {
			defer s.wg.Done()
			if err := s.proxy.Shutdown(shutdownCtx); err != nil {
				slog.Error("Proxy server shutdown error", "error", err)
			}
		}()

		s.wg.Wait()
	}()

	errCh := make(chan error, 2)

	go func() {
		if err := s.dashboard.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("dashboard server error: %w", err)
			return
		}
		errCh <- nil
	}()

	go func() {
		if err := s.proxy.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("proxy server error: %w", err)
			return
		}
		errCh <- nil
	}()

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
