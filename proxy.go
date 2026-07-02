package main

import (
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

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
func IsFromDashboard(r *http.Request) bool {
	cookie, err := r.Cookie("telescope_dashboard")
	return err == nil && cookie.Value == "1"
}

func (s *Server) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// if IsFromDashboard(r) {
	// 	w.WriteHeader(200)
	// }
	host, targetPath := GetHostAndPath(r)
	targetAddress, found := s.routeTable.Lookup(host, targetPath)
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
