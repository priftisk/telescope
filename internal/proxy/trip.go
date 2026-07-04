package proxy

import (
	"net/http"
	"time"
)

type Trip struct {
	req      *http.Request
	resp     *http.Response
	duration time.Duration
}
