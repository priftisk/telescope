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

func MakeAndServe(targetURL *url.URL, targetPath string, w http.ResponseWriter, r *http.Request) {

	proxy := &httputil.ReverseProxy{
		Rewrite: RewriteProxy(targetURL, targetPath, r),
	}

	proxy.ServeHTTP(w, r)
	log.Printf("PROXY %s %s %s → %s",
		r.Method, r.URL.Path, r.Host, targetURL)
}

// func IsFromDashboard(r *http.Request) bool {
// 	cookie, err := r.Cookie("telescope_dashboard")
// 	return err == nil && cookie.Value == "1"
// }
