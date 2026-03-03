package proxy

import (
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"turnstile/internal/auth"
)

// idempotentMethods is the set of HTTP methods that are safe to retry
// without risk of duplicate side effects on the upstream.
var idempotentMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
	http.MethodPut:     true,
	http.MethodDelete:  true,
}

// retryTransport wraps an http.RoundTripper and retries requests on
// connection-level errors for idempotent HTTP methods.
type retryTransport struct {
	wrapped    http.RoundTripper
	maxRetries int
	baseDelay  time.Duration
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !idempotentMethods[req.Method] {
		return t.wrapped.RoundTrip(req)
	}

	var (
		resp *http.Response
		err  error
	)

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		resp, err = t.wrapped.RoundTrip(req)
		if err == nil {
			return resp, nil
		}

		// Only retry on connection-level errors (dial failures, connection
		// refused, etc.) — not on successful HTTP responses, even 5xx, since
		// those mean the upstream received the request.
		if !isConnectionError(err) {
			return nil, err
		}

		if attempt < t.maxRetries {
			delay := t.backoff(attempt)
			slog.Debug("retrying upstream",
				"method", req.Method,
				"url", req.URL.String(),
				"attempt", attempt+1,
				"max_retries", t.maxRetries,
				"delay", delay,
				"error", err,
			)
			time.Sleep(delay)
		}
	}

	return nil, err
}

// backoff returns an exponential backoff duration capped at 10 seconds.
func (t *retryTransport) backoff(attempt int) time.Duration {
	multiplier := math.Pow(2, float64(attempt))
	delay := time.Duration(float64(t.baseDelay) * multiplier)
	const maxDelay = 10 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// isConnectionError reports whether err is a connection-level failure
// (dial error, connection refused, etc.) rather than a higher-level error.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	urlErr, ok := err.(*url.Error)
	if !ok {
		return false
	}
	_, isNetErr := urlErr.Err.(*net.OpError)
	return isNetErr
}

type Handler struct {
	reverseProxy *httputil.ReverseProxy
}

func NewHandler(backendURL string, maxRetries int, retryDelay time.Duration) (*Handler, error) {
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

		// Resolve and log the upstream IP(s) at debug level. This runs a fresh
		// DNS lookup on every request, so it reflects any IP changes due to
		// Railway service sleep/wake cycles (no in-process DNS caching).
		if addrs, err := net.LookupHost(target.Hostname()); err == nil {
			slog.Debug("upstream resolved", "host", target.Host, "ips", addrs)
		}
	}

	// Flush immediately: fixes SSE workloads
	proxy.FlushInterval = -1

	proxy.Transport = &retryTransport{
		wrapped:    http.DefaultTransport,
		maxRetries: maxRetries,
		baseDelay:  retryDelay,
	}

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
