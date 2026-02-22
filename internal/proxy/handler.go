package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
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
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			// manual IP lookup to support legacy env
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, fmt.Errorf("dns lookup %q: %w", host, err)
			}

			// Sort so IPv6 addresses are tried first
			sort.Slice(ips, func(i, j int) bool {
				return ips[i].IP.To4() == nil && ips[j].IP.To4() != nil
			})

			d := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}

			var lastErr error
			for _, ip := range ips {
				conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(ip.IP.String(), port))
				if err == nil {
					return conn, nil
				}
				lastErr = err
			}
			return nil, fmt.Errorf("all addresses failed for %q (tried %d): %w", host, len(ips), lastErr)
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

	// Flush immediately: fixes SSE workloads
	proxy.FlushInterval = -1

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("proxy error", "method", r.Method, "path", r.URL.Path, "error", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}

	return &Handler{reverseProxy: proxy}, nil
}

func (p *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.reverseProxy.ServeHTTP(w, r)
}
