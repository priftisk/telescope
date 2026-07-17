package proxy

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"telescope/internal/config"
	"telescope/internal/roundtripper"
	"telescope/internal/router"
)

type ProxyServer struct {
	config       *config.ProxyConfig
	routeTable   *router.RouteTable
	trips        *roundtripper.Trips
	httpServer   *http.Server
	roundTripper *roundtripper.ProxyRoundTripper
}

func NewProxyServer(rt *router.RouteTable, trips *roundtripper.Trips, opts ...config.Opt) *ProxyServer {
	p := &ProxyServer{
		config:       &config.ProxyConfig{},
		routeTable:   rt,
		roundTripper: roundtripper.NewProxyRoundTripper(trips),
	}
	for _, opt := range opts {
		opt(p.config)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.ProxyHandler)

	p.httpServer = &http.Server{
		Addr:    p.config.Host + ":" + p.config.Port,
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
	slog.Info("Telescope proxy listening on", "addr", "http://"+p.config.Host+":"+p.config.Port)
	return p.httpServer.ListenAndServe()
}

func (p *ProxyServer) Shutdown(ctx context.Context) error {
	return p.httpServer.Shutdown(ctx)
}
