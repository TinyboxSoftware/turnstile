package proxy

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"turnstile/internal/auth"
)

type Handler struct {
	reverseProxy *httputil.ReverseProxy
}

func NewHandler(backendURL string) (*Handler, error) {
	target, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {

			// Setup a basic dialer; mirrors the default from what I can tell
			dialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}

			// Try IPv6 first then fallback to IPv4 (legacy Railway internal networking is IPv6 only)
			if conn, err := dialer.DialContext(ctx, "tcp6", addr); err == nil {
				return conn, nil
			}
			return dialer.DialContext(ctx, "tcp4", addr)
		},

		// more defaults from the base implementation
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	proxy.Director = func(req *http.Request) {
		originalHost := req.Host

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		if req.Header.Get("X-Forwarded-Host") == "" {
			req.Header.Set("X-Forwarded-Host", originalHost)
		}
		if req.Header.Get("X-Forwarded-Proto") == "" {
			req.Header.Set("X-Forwarded-Proto", "https")
		}

		session := auth.GetSessionFromContext(req.Context())
		if session != nil {
			req.Header.Set("X-Auth-Email", session.Email)
			req.Header.Set("X-Auth-User-ID", session.UserID)
			req.Header.Set("X-Auth-Name", session.Name)
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}

	return &Handler{reverseProxy: proxy}, nil
}

func (p *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.reverseProxy.ServeHTTP(w, r)
}
