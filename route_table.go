package main

import (
	"log/slog"
	"slices"
	"strings"
	"sync"
)

type RouteTable struct {
	Routes    []Route
	Mutex     sync.RWMutex
	HostIndex map[string][]*Route
}

func (rt *RouteTable) Register(container ContainerInfo) {
	rt.Mutex.Lock()
	defer rt.Mutex.Unlock()
	labels := container.Labels
	rt.Routes = append(rt.Routes, Route{
		HostName:      labels.ProxyHost,
		TargetAddress: container.ContainerIPAddr + ":" + labels.ProxyPort,
		URLPath:       labels.ProxyPath,
	})
	slog.Info("registered route",
		"host", labels.ProxyHost,
		"addr", labels.ProxyPort,
	)
}
func (rt *RouteTable) Deregister(host string) {
	rt.Mutex.Lock()
	rt.Routes = slices.DeleteFunc(rt.Routes, func(r Route) bool {
		return r.HostName == host
	})
	slog.Info("deregistering container route", "container", host)
	defer rt.Mutex.Unlock()
}

func (rt *RouteTable) Lookup(host string, path string) (string, bool) {
	rt.Mutex.RLock()
	defer rt.Mutex.RUnlock()

	// Strip port from host for comparison (e.g., "localhost:8901" -> "localhost")
	hostOnly := StripPort(host)

	for _, route := range rt.Routes {

		if !HostMatches(hostOnly, route.HostName) {
			continue
		}

		// Match path (if route has a path specified)
		if route.URLPath != "" && route.URLPath != "/" {
			if !strings.HasPrefix(path, route.URLPath) {
				continue
			}
		}

		return route.TargetAddress, true
	}

	return "", false
}

func (rt *RouteTable) List() []Route {
	rt.Mutex.RLock()
	defer rt.Mutex.RUnlock()
	return rt.Routes
}
