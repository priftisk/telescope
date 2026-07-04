package dashboard

import (
	"telescope/internal/router"
	"time"
)

type DashboardData struct {
	Routes      []router.Route `json:"routes"`
	TotalRoutes int            `json:"total_routes"`
	Uptime      int64          `json:"uptime"`
	Version     string         `json:"version"`
}

func NewDashboardData(routeTable *router.RouteTable, upTime time.Time) *DashboardData {
	return &DashboardData{
		Routes:      routeTable.Routes,
		TotalRoutes: len(routeTable.Routes),
		Uptime:      time.Since(upTime).Milliseconds(),
		Version:     "1.0.0",
	}
}
