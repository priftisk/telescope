package server

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"telescope/internal"
	"telescope/internal/dashboard"
	"telescope/internal/proxy"
	router "telescope/internal/router"
	"text/template"
	"time"
)

func (s *Server) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// if IsFromDashboard(r) {
	// 	w.WriteHeader(200)
	// }
	host, targetPath := router.GetHostAndPath(r)
	targetAddress, found := s.routeTable.Lookup(host, targetPath)
	if !found {
		log.Printf("Proxy fail: %s | %s not found\n", host, targetPath)
		w.WriteHeader(502)
		return
	}

	targetURL, err := url.Parse("http://" + targetAddress)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(500)
		return
	}
	proxy.MakeAndServe(targetURL, targetPath, w, r)
}

func (s *Server) DashboardHandler(w http.ResponseWriter, r *http.Request) {

	data := dashboard.DashboardData{
		Routes:      s.routeTable.Routes,
		TotalRoutes: len(s.routeTable.Routes),
		Uptime:      time.Since(s.startTime),
		Version:     "1.0.0",
	}
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
