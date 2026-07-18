package dashboard

import (
	"net"
	"telescope/internal/config"
	"telescope/internal/roundtripper"
	"telescope/internal/router"
	"time"
)

type DashboardData struct {
	Routes      []router.Route `json:"routes"`
	TotalRoutes int            `json:"total_routes"`
	Uptime      int64          `json:"uptime"`
	ProxyAddr   string         `json:"proxy_addr"`
	Version     string         `json:"version"`
}

func NewDashboardData(routeTable *router.RouteTable, upTime time.Time, conf *config.ServerConfig) *DashboardData {
	return &DashboardData{
		Routes:      routeTable.Routes,
		TotalRoutes: len(routeTable.Routes),
		Uptime:      time.Since(upTime).Milliseconds(),
		ProxyAddr:   net.JoinHostPort(conf.ProxyHost, conf.ProxyPort),
		Version:     "1.0.0",
	}

}

type TripsData struct {
	Trips    []roundtripper.Trip
	TripsNum int64
}
