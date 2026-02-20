package auth

import (
	"net/http"
	"strings"

	"railway-oauth-proxy/internal/httpx"
	"railway-oauth-proxy/internal/session"
)

type Middleware struct {
	session *session.Manager
}

func NewMiddleware(sessionManager *session.Manager) *Middleware {
	return &Middleware{session: sessionManager}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := m.session.GetSession(r)
		if err != nil {
			httpx.WriteJSONError(w, "session_error", "Invalid session. Please log in again.", http.StatusUnauthorized)
			return
		}

		if sess == nil {
			if isAPIRequest(r) {
				httpx.WriteJSONError(w, "unauthorized", "Session expired. Please log in again.", http.StatusUnauthorized)
				return
			}
			loginURL := "/oauth/login"
			if r.URL.Path != "/" {
				loginURL += "?redirect=" + r.URL.RequestURI()
			}
			http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
			return
		}

		r = r.WithContext(SetSessionContext(r.Context(), sess))
		next.ServeHTTP(w, r)
	})
}

func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") ||
		r.Header.Get("Accept") == "application/json" ||
		r.Header.Get("Content-Type") == "application/json"
}
