package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"telescope/internal/router"
)

func (p *ProxyServer) MakeAndServe(route *router.Route, targetURL *url.URL, targetPath string, w http.ResponseWriter, r *http.Request) {

	proxy := &httputil.ReverseProxy{
		Rewrite:   RewriteProxy(route, targetURL, targetPath, r),
		Transport: p.roundTripper,
	}

	proxy.ServeHTTP(w, r)
	log.Printf("PROXY %s %s %s → %s",
		r.Method, r.URL.Path, r.Host, targetURL)
}

func RewriteProxy(route *router.Route, targetURL *url.URL, targetPath string, r *http.Request) func(*httputil.ProxyRequest) {
	return func(pr *httputil.ProxyRequest) {

		pr.SetURL(targetURL)
		pr.Out.Host = r.Host
		if targetPath != "" {
			if route.URLPath == "/" { // Registered route has no path label, so dont strip anything
				pr.Out.URL.Path = r.URL.Path

			} else { // Strip the routing prefix before forwarding
				pr.Out.URL.Path = strings.TrimPrefix(r.URL.Path, "/"+targetPath)

			}
			if pr.Out.URL.Path == "" {
				pr.Out.URL.Path = "/"
			}
			pr.Out.URL.RawPath = ""
		}

	}
}
