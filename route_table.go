package main

import (
	"log/slog"
	"slices"
	"sync"
)

type RouteTable struct {
	Routes []Route
	Mutex  sync.RWMutex
}

func (rt *RouteTable) Register(labels Labels, containerIP string) {
	rt.Mutex.Lock()
	defer rt.Mutex.Unlock()

	rt.Routes = append(rt.Routes, Route{
		HostName:      labels.ProxyHost,
		TargetAddress: containerIP + ":" + labels.ProxyPort,
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
	for _, route := range rt.Routes {
		slog.Info("Checking", "host", host, "proxhost", route.HostName)
		if route.URLPath == path {
			return route.TargetAddress, true
		}
	}
	return "", false
}

func (rt *RouteTable) List() []Route {
	rt.Mutex.RLock()
	defer rt.Mutex.RUnlock()
	return rt.Routes
}
