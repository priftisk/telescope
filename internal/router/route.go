package router

import (
	"net"
	"net/http"
	"strings"
	"telescope/internal/container"
)

type Route struct {
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	HostName      string `json:"hostname"`
	TargetAddress string `json:"address"`
	URLPath       string `json:"url_path"`
}

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

func NewRoute(container container.ContainerInfo) Route {

	return Route{
		ContainerID:   container.ContainerID,
		ContainerName: container.ContainerName,
		HostName:      container.Labels.ProxyHost,
		TargetAddress: container.ContainerIPAddr + ":" + container.Labels.ProxyPort,
		URLPath:       container.Labels.ProxyPath,
	}
}
func PatternHostMatches(requestHost, routeHost string) bool {

	// Wildcard match ("*.example.com")
	if strings.HasPrefix(routeHost, "*.") {
		domain := routeHost[2:] // Remove "*."
		return strings.HasSuffix(requestHost, "."+domain) || requestHost == domain
	}

	return false
}

func pathMatches(requestPath, routePath string) bool {
	// Exact match
	if requestPath == routePath {
		return true
	}

	if routePath == "/" {
		return true
	}

	if !strings.HasPrefix(requestPath, routePath) {
		return false
	}

	return len(requestPath) == len(routePath) || // Paths are same
		requestPath[len(routePath)] == '/' // Or next char after path is "/"

}

func HostMatches(requestHost, routeHost string) bool {
	if requestHost == routeHost {
		return true
	}
	return PatternHostMatches(requestHost, routeHost)
}
