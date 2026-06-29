package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func runProxy(ctx context.Context, rt *RouteTable) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		targetHost := r.Host
		targetAddress, found := rt.Lookup(targetHost)
		if !found {
			w.WriteHeader(502)
			return
		}
		targetURL, err := url.Parse("http://" + targetAddress)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.ServeHTTP(w, r)
		slog.Info("proxied: %s -> %s", targetHost, targetAddress)
	})

	server := &http.Server{
		Addr:    ":8901",
		Handler: mux,
	}

	// watch for ctx cancellation
	go func() {
		<-ctx.Done()
		slog.Info("context cancelled, shutting down proxy")
		server.Shutdown(context.Background())
	}()

	slog.Info("Proxy listening on :8901")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("proxy error: %v", err)
	}
}
