package proxy

import (
	"net/http"
	"time"
)

type Trip struct {
	RequestMethod string `json:"method"`
	RequestPath   string `json:"path"`
	RequestQuery  string `json:"query,omitempty"`
	RequestHost   string `json:"host"`
	RemoteAddr    string `json:"remote_addr,omitempty"`
	UserAgent     string `json:"user_agent,omitempty"`

	ResponseStatus     string `json:"response_status"`
	ResponseStatusCode int    `json:"status_code"`
	ResponseSize       int64  `json:"response_size_bytes"`

	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`

	Error string `json:"error,omitempty"`
}

func NewTrip(req *http.Request, resp *http.Response, duration time.Duration) Trip {
	trip := Trip{
		RequestMethod: req.Method,
		RequestPath:   req.URL.Path,
		RequestQuery:  req.URL.RawQuery,
		RequestHost:   req.Host,
		RemoteAddr:    req.RemoteAddr,
		UserAgent:     req.UserAgent(),
		Duration:      duration,
		Timestamp:     time.Now(),
	}

	if resp != nil {
		trip.ResponseStatus = resp.Status
		trip.ResponseStatusCode = resp.StatusCode
		trip.ResponseSize = resp.ContentLength
	}

	return trip
}
