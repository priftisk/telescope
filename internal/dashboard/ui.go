package dashboard

import (
	"telescope/internal/router"
	"time"
)

type DashboardData struct {
	Routes      []router.Route
	TotalRoutes int
	Uptime      time.Duration
	Version     string
}
