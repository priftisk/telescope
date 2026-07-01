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

func StripPort(host string) string {
	hostOnly, _, err := net.SplitHostPort(host)
	if err != nil {
		// No port present.
		return host
	}
	return hostOnly
}

func GetHostAndPath(r *http.Request) (string, string) {
	host := StripPort(r.Host)

	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		return host, "/"
	}

	if i := strings.IndexByte(path, '/'); i != -1 {
		path = path[:i]
	}

	return host, path
}

func RewriteProxy(targetURL *url.URL, targetPath string, r *http.Request) func(*httputil.ProxyRequest) {
	return func(pr *httputil.ProxyRequest) {

		pr.SetURL(targetURL)
		pr.Out.Host = r.Host

		// Strip the routing prefix before forwarding
		if targetPath != "" {
			pr.Out.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+targetPath)
			if pr.Out.URL.Path == "" {
				pr.Out.URL.Path = "/"
			}
			pr.Out.URL.RawPath = ""
		}

	}
}

func MakeAndServe(targetURL *url.URL, targetPath string, w http.ResponseWriter, r *http.Request) {
	proxy := &httputil.ReverseProxy{
		Rewrite: RewriteProxy(targetURL, targetPath, r),
	}

	proxy.ServeHTTP(w, r)
	log.Printf("PROXY %s %s %s → %s",
		r.Method, r.URL.Path, r.Host, targetURL)
}

func ProxyHandler(rt *RouteTable) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		host, targetPath := GetHostAndPath(r)
		targetAddress, found := rt.Lookup(host, targetPath)
		if !found {
			log.Printf("Proxy fail: %s | %s not found\n", host, targetPath)
			w.WriteHeader(502)
			return
		}

		targetURL, err := url.Parse("http://" + targetAddress)
		if err != nil {
			slog.Error(err.Error())
			w.WriteHeader(500)
			return
		}
		MakeAndServe(targetURL, targetPath, w, r)
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
