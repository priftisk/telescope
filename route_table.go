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

func (rt *RouteTable) Register(container ContainerInfo) {
	rt.Mutex.Lock()
	defer rt.Mutex.Unlock()
	labels := container.Labels
	rt.Routes = append(rt.Routes, Route{
		HostName:      labels.ProxyHost,
		TargetAddress: container.ContainerID + ":" + labels.ProxyPort,
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
		// Cannot directly check for host == rt.HostName, because
		// host will be localhost:XXXX and rt.HostName will be localhost
		if isLocal := IsLocalhost(host); isLocal == true {
			if route.URLPath == path && route.HostName == "localhost" {
				return route.TargetAddress, true
			}
		} else {
			if route.URLPath == path && route.HostName == host {
				return route.TargetAddress, true
			}
		}
	}
	return "", false
}

func (rt *RouteTable) List() []Route {
	rt.Mutex.RLock()
	defer rt.Mutex.RUnlock()
	return rt.Routes
}
