package proxy

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ProxyRoundTripper struct {
	Transport http.RoundTripper
	Trips     *Trips
	mu        sync.Mutex
}

func NewProxyRoundTripper(trips *Trips) *ProxyRoundTripper {
	return &ProxyRoundTripper{
		Transport: http.DefaultTransport,
		Trips:     trips,
	}
}

func (pt *ProxyRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	transport := pt.Transport
	start := time.Now()
	resp, err := transport.RoundTrip(r)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return nil, err
	}
	elapsed := time.Since(start)
	// Record the trip
	trip := Trip{
		req:      r,
		resp:     resp,
		duration: elapsed,
	}
	pt.Trips.Add(trip)

	return resp, nil
}
