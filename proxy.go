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

func StripPort(host string) string {
	hostOnly, _, err := net.SplitHostPort(host)
	if err != nil {
		// If no port, SplitHostPort returns an error
		return host
	}
	return hostOnly
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

func ProxyHandler(rt *RouteTable) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		Host, targetPath := GetHostAndPath(r)
		targetAddress, found := rt.Lookup(Host, targetPath)
		if !found {
			log.Printf("Proxy fail: %s not found\n", targetAddress)
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
		log.Printf("PROXY %s %s %s → %s",
			r.Method, r.URL.Path, r.Host, targetAddress)
	}
}

func RoutesHandler(rt *RouteTable) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		routes := rt.List()
		// fmt.Println(routes)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(routes)
	}
}

func RunProxy(ctx context.Context, rt *RouteTable) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", ProxyHandler(rt))
	mux.HandleFunc("GET /routes", RoutesHandler(rt))

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
