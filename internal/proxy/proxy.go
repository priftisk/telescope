package proxy

import (
	"log"
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

func (p *ProxyServer) MakeAndServe(targetURL *url.URL, targetPath string, w http.ResponseWriter, r *http.Request) {

	proxy := &httputil.ReverseProxy{
		Rewrite:   RewriteProxy(targetURL, targetPath, r),
		Transport: p.roundTripper,
	}

	proxy.ServeHTTP(w, r)
	log.Printf("PROXY %s %s %s → %s",
		r.Method, r.URL.Path, r.Host, targetURL)
}
