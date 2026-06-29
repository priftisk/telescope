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

func (rt *RouteTable) Register(host string, target string) {
	rt.Mutex.Lock()
	defer rt.Mutex.Unlock()

	rt.Routes = append(rt.Routes, Route{
		HostName:      host,
		TargetAddress: target,
	})
	slog.Info("registered route",
		"host", host,
		"addr", target,
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

func (rt *RouteTable) Lookup(host string) (string, bool) {
	rt.Mutex.RLock()
	defer rt.Mutex.RUnlock()
	for _, route := range rt.Routes {
		if route.HostName == host {
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
