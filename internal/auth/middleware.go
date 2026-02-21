package auth

import (
	"net/http"
	"net/url"
	"strings"

	"turnstile/internal/httpx"
	"turnstile/internal/session"
)

type Middleware struct {
	session   *session.Manager
	loginPath string
}

func NewMiddleware(sessionManager *session.Manager, loginPath string) *Middleware {
	return &Middleware{session: sessionManager, loginPath: loginPath}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := m.session.GetSession(r)
		if err != nil {
			httpx.WriteJSONError(w, "session_error", "Invalid session. Please log in again.", http.StatusUnauthorized)
			return
		}

		if sess == nil {

			// if this is an API request, don't redirect, just 401
			if isAPIRequest(r) {
				httpx.WriteJSONError(w, "unauthorized", "Session expired. Please log in again.", http.StatusUnauthorized)
				return
			}

			// if not an API request, redirect the user and log them in
			loginURL := m.loginPath
			if r.URL.Path != "/" {
				loginURL += "?redirect=" + url.QueryEscape(r.URL.RequestURI())
			}
			http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
			return
		}

		r = r.WithContext(SetSessionContext(r.Context(), sess))
		next.ServeHTTP(w, r)
	})
}

// try to be smart and determine if this is an API request
func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") ||
		r.Header.Get("Accept") == "application/json" ||
		r.Header.Get("Content-Type") == "application/json"
}
