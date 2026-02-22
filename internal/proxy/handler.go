package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

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
