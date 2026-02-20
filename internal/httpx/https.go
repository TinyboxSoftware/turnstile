package httpx

import (
	"net/http"
	"strings"
)

func IsHTTPS(r *http.Request) bool {
	if r.URL.Scheme == "https" {
		return true
	}
	if r.TLS != nil {
		return true
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return strings.EqualFold(proto, "https")
	}
	return false
}
