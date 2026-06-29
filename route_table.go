package main

import (
	"log"
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
	log.Printf("Registered: %s -> %s", host, target)
}
func (rt *RouteTable) Deregister(host string) {
	rt.Mutex.Lock()
	rt.Routes = slices.DeleteFunc(rt.Routes, func(r Route) bool {
		return r.HostName == host
	})
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
