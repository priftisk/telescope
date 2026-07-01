package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func MakeAndServe(targetURL *url.URL, w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(w, r)
}

func IsLocalhost(host string) bool {
	host, _, _ = strings.Cut(host, ":")

	if host == "localhost" {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func GetHostAndPath(r *http.Request) (string, string) {
	var targetHost, targetPath string
	targetHost = r.Host

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 {
		targetPath = ""
	} else {
		targetPath = parts[0]
	}
	return targetHost, targetPath
}

func runProxy(ctx context.Context, rt *RouteTable) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Host, targetPath := GetHostAndPath(r)
		targetAddress, found := rt.Lookup(Host, targetPath)
		if !found {
			w.WriteHeader(502)
			return
		}

		targetURL, err := url.Parse("http://" + targetAddress)
		if err != nil {
			slog.Error(err.Error())
			w.WriteHeader(500)
			return
		}

		if targetPath != "" {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+targetPath)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			r.URL.RawPath = "" // avoid stale escaped path overriding Path
		}

		MakeAndServe(targetURL, w, r)
		slog.Info("proxied", "host", r.Host, "address", targetAddress, "path", r.URL.Path)
	})
	mux.HandleFunc("GET /routes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		routes := rt.List()
		// fmt.Println(routes)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(routes)
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
