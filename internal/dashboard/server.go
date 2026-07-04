package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"telescope/internal"
	"telescope/internal/router"
	"text/template"
	"time"
)

type DashboardServer struct {
	routeTable *router.RouteTable
	startTime  time.Time
	httpServer *http.Server
}

func NewDashboardServer(rt *router.RouteTable, startTime time.Time) (*DashboardServer, error) {
	d := &DashboardServer{
		routeTable: rt,
		startTime:  startTime,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /routes", d.RoutesHandler)
	mux.HandleFunc("GET /dashboard", d.DashboardHandler)
	mux.HandleFunc("/dashboard/{resource}", d.DashboardResourceHandler)

	staticDir, err := internal.GetStaticDir()
	if err != nil {
		return nil, err
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(staticDir))))

	d.httpServer = &http.Server{
		Addr:    ":8900",
		Handler: mux,
	}
	return d, nil
}

func (d *DashboardServer) RoutesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	routes := d.routeTable.List()
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(routes)
}

func (d *DashboardServer) DashboardHandler(w http.ResponseWriter, r *http.Request) {

	data := NewDashboardData(d.routeTable, d.startTime)
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
	data := NewDashboardData(d.routeTable, d.startTime)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (d *DashboardServer) ListenAndServe() error {
	return d.httpServer.ListenAndServe()
}

func (d *DashboardServer) Shutdown(ctx context.Context) error {
	return d.httpServer.Shutdown(ctx)
}
