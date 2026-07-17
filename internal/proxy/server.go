package proxy

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"telescope/internal/roundtripper"
	"telescope/internal/router"
)

type ProxyServer struct {
	routeTable   *router.RouteTable
	trips        *roundtripper.Trips
	httpServer   *http.Server
	roundTripper *roundtripper.ProxyRoundTripper
}

func NewProxyServer(rt *router.RouteTable, trips *roundtripper.Trips) *ProxyServer {
	p := &ProxyServer{
		routeTable:   rt,
		roundTripper: roundtripper.NewProxyRoundTripper(trips),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.ProxyHandler)

	p.httpServer = &http.Server{
		Addr:    ":8901",
		Handler: mux,
	}
	return p
}

func (proxy *ProxyServer) ProxyHandler(w http.ResponseWriter, r *http.Request) {

	host, targetPath := router.GetHostAndPath(r)
	route, found := proxy.routeTable.Lookup(host, targetPath)
	if !found {
		log.Printf("Proxy fail: %s | %s not found\n", host, targetPath)
		w.WriteHeader(502)
		return
	}

	targetURL, err := url.Parse("http://" + route.TargetAddress)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(500)
		return
	}
	proxy.MakeAndServe(route, targetURL, targetPath, w, r)
}

func (p *ProxyServer) ListenAndServe() error {
	return p.httpServer.ListenAndServe()
}

func (p *ProxyServer) Shutdown(ctx context.Context) error {
	return p.httpServer.Shutdown(ctx)
}
