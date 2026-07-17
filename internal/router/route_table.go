package router

import (
	"log/slog"
	"slices"
	"sync"
	"telescope/internal/container"
)

type RouteTable struct {
	Routes    []Route
	Mutex     sync.RWMutex
	HostIndex map[string][]*Route
}

func NewRouteTable() *RouteTable {
	return &RouteTable{
		Routes:    make([]Route, 0),
		HostIndex: map[string][]*Route{},
	}
}

func (rt *RouteTable) Register(container container.ContainerInfo) {
	rt.Mutex.Lock()
	defer rt.Mutex.Unlock()

	new_route := NewRoute(container)
	rt.Routes = append(rt.Routes, new_route)

	if rt.HostIndex == nil {
		rt.HostIndex = make(map[string][]*Route)
	}

	lastIdx := len(rt.Routes) - 1
	rt.HostIndex[new_route.HostName] = append(rt.HostIndex[new_route.HostName], &rt.Routes[lastIdx])

	slog.Info("Registered", "container", container.ContainerName, "host", container.Labels.ProxyHost, "addr", container.Labels.ProxyPort)

}

func (rt *RouteTable) Deregister(containerID string) {
	rt.Mutex.Lock()
	rt.Routes = slices.DeleteFunc(rt.Routes, func(r Route) bool {
		return r.ContainerID == containerID
	})

	slog.Info("Deregistered", "container", containerID)
	defer rt.Mutex.Unlock()
}

func (rt *RouteTable) Lookup(host string, path string) (*Route, bool) {
	rt.Mutex.RLock()
	defer rt.Mutex.RUnlock()

	// Strip port from host for comparison  ("localhost:8901" -> "localhost")
	hostOnly := StripPort(host)

	candidates := rt.HostIndex[hostOnly]

	if len(candidates) == 0 { // Not exact host match so check if host matches pattern
		for host, routes := range rt.HostIndex {
			if HostMatches(hostOnly, host) {
				candidates = routes
				break
			}
		}
	}

	var bestMatch *Route
	bestPathLen := -1

	for _, route := range candidates { // Match based on closest path length
		routePath := route.URLPath
		if routePath == "" {
			routePath = "/"
		}

		if pathMatches(path, routePath) && len(routePath) > bestPathLen {
			bestMatch = route
			bestPathLen = len(routePath)
		}
	}

	if bestMatch != nil {
		return bestMatch, true
	}

	return &Route{}, false
}

func (rt *RouteTable) List() []Route {
	rt.Mutex.RLock()
	defer rt.Mutex.RUnlock()
	return rt.Routes
}
