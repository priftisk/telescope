package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"telescope/internal"
	"telescope/internal/config"
	"telescope/internal/roundtripper"
	"telescope/internal/router"
	"text/template"
	"time"
)

type DashboardServer struct {
	config     *config.ServerConfig
	routeTable *router.RouteTable
	trips      *roundtripper.Trips
	httpServer *http.Server
	startTime  time.Time
}

func (d *DashboardServer) ListenAndServe() error {
	slog.Info("Dashboard available at", "addr", "http://"+net.JoinHostPort(d.config.DashboardHost, d.config.DashboardPort)+"/dashboard")
	return d.httpServer.ListenAndServe()
}

func (d *DashboardServer) Shutdown(ctx context.Context) error {
	return d.httpServer.Shutdown(ctx)
}

func (d *DashboardServer) applyOpts(opts ...config.Opt) {
	for _, opt := range opts { // Apply opts
		opt(d.config)
	}
}

func NewDashboardServer(rt *router.RouteTable, trips *roundtripper.Trips, startTime time.Time, opts ...config.Opt) (*DashboardServer, error) {
	d := &DashboardServer{
		config: &config.ServerConfig{
			DashboardHost: config.DefaultHost,
			DashboardPort: config.DefaultDashboardPort,
		}, // Defaults
		routeTable: rt,
		trips:      trips,
		startTime:  startTime,
	}
	d.applyOpts(opts...)
	mux := http.NewServeMux()
	// mux.HandleFunc("GET /routes", d.RoutesHandler)
	mux.HandleFunc("GET /dashboard", d.DashboardHandler)
	mux.HandleFunc("GET /dashboard/{resource}", d.DashboardResourceHandler)

	staticDir, err := internal.GetStaticDir()
	if err != nil {
		return nil, err
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(staticDir))))
	d.httpServer = &http.Server{
		Addr:    net.JoinHostPort(d.config.DashboardHost, d.config.DashboardPort),
		Handler: mux,
	}
	return d, nil
}

func (d *DashboardServer) DashboardHandler(w http.ResponseWriter, r *http.Request) {

	data := NewDashboardData(d.routeTable, d.startTime, d.config)
	static_dir, err := internal.GetStaticDir()
	if err != nil {
		fmt.Printf("%+v\n", err.Error())
		w.WriteHeader(500)
		return
	}
	tmpl, err := template.ParseFiles(filepath.Join(static_dir, "dashboard.html"))

	if err != nil {
		fmt.Printf("%+v\n", err.Error())
		w.WriteHeader(500)
		return
	}
	tmpl.Execute(w, data)
}

func (d *DashboardServer) DashboardResourceHandler(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	switch resource {
	case "data":
		data := NewDashboardData(d.routeTable, d.startTime, d.config)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(data)
	case "trips":
		trips := d.trips.GetAll()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(trips)
	default:
		w.WriteHeader(404)
		json.NewEncoder(w).Encode([]byte(resource + " not found."))
	}

}

func (d *DashboardServer) RoutesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	routes := d.routeTable.List()
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(routes)
}
