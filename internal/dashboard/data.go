package dashboard

import (
	"telescope/internal/router"
)

type DashboardData struct {
	Routes      []router.Route `json:"routes"`
	TotalRoutes int            `json:"total_routes"`
	Uptime      int64          `json:"uptime"`
	Version     string         `json:"version"`
}
