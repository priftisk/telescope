package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type DashboardData struct {
	Routes      []Route
	TotalRoutes int
	Uptime      string
	Version     string
}

func (s *Server) DashboardHandler(w http.ResponseWriter, r *http.Request) {

	data := DashboardData{
		Routes:      s.routeTable.Routes,
		TotalRoutes: len(s.routeTable.Routes),
		Uptime:      time.Since(s.startTime).String(),
		Version:     "1.0.0",
	}
	tmpl, err := template.ParseFiles("./static/dashboard.html")
	if err != nil {
		fmt.Printf("%+v\n", err.Error())
		w.WriteHeader(500)
		return
	}
	tmpl.Execute(w, data)
}
