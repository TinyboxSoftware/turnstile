package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"turnstile/internal/config"
	"turnstile/internal/httpx"
	"turnstile/internal/railway"
	"turnstile/internal/session"
	"turnstile/internal/views"
)

const (
	oauthAuthURL       = "https://backboard.railway.com/oauth/auth"
	oauthTokenURL      = "https://backboard.railway.com/oauth/token"
	redirectCookieName = "oauth_redirect"
)

type Handler struct {
	cfg      *config.Config
	session  *session.Manager
	railway  *railway.Client
	renderer *views.Renderer
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
	Scope       string `json:"scope"`
}

func NewHandler(cfg *config.Config, sessionManager *session.Manager, railwayClient *railway.Client, renderer *views.Renderer) *Handler {
	return &Handler{
		cfg:      cfg,
		session:  sessionManager,
		railway:  railwayClient,
		renderer: renderer,
	}
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// If this is an error redirect from a failed auth attempt, show an error page
	// rather than immediately re-initiating OAuth (which would loop).
	if errType := r.URL.Query().Get("error"); errType != "" {
		h.handleAuthError(w, r, errType)
		return
	}

	state, err := generateState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   httpx.IsHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})

	// Persist the post-login redirect destination in a cookie
	if redirectTo := r.URL.Query().Get("redirect"); isSafeRedirect(redirectTo) {
		http.SetCookie(w, &http.Cookie{
			Name:     redirectCookieName,
			Value:    redirectTo,
			Path:     "/",
			MaxAge:   600,
			HttpOnly: true,
			Secure:   httpx.IsHTTPS(r),
			SameSite: http.SameSiteLaxMode,
		})
	}

	redirectURI := h.cfg.URI(config.RouteCallback, config.FullURL)
	slog.Info("oauth_login", "redirect_uri", redirectURI, "is_https", httpx.IsHTTPS(r))

	params := url.Values{
		"response_type": {"code"},
		"client_id":     {h.cfg.RailwayClientID},
		"redirect_uri":  {redirectURI},
		"scope":         {"openid email profile project:viewer"},
		"state":         {state},
	}

	if r.URL.Query().Get("reconsent") == "true" {
		params.Set("prompt", "consent")
	}

	authURL := oauthAuthURL + "?" + params.Encode()
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *Handler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("oauth_callback", "query", r.URL.RawQuery)

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		slog.Error("oauth_callback_error", "error", "missing_state_cookie", "err", err)
		httpx.WriteJSONError(w, "invalid_request", "Missing state cookie", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" || state != stateCookie.Value {
		slog.Error("oauth_callback_error", "error", "invalid_state", "state", state, "cookie", stateCookie.Value)
		httpx.WriteJSONError(w, "invalid_request", "Invalid state parameter", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   httpx.IsHTTPS(r),
	})

	// Read and immediately clear the redirect cookie so it isn't reused.
	redirectURL := "/"
	if redirectCookie, cookieErr := r.Cookie(redirectCookieName); cookieErr == nil {
		if isSafeRedirect(redirectCookie.Value) {
			redirectURL = redirectCookie.Value
		}
		http.SetCookie(w, &http.Cookie{
			Name:     redirectCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   httpx.IsHTTPS(r),
		})
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		errorDesc := r.URL.Query().Get("error_description")
		if errorDesc == "" {
			errorDesc = r.URL.Query().Get("error")
		}
		if errorDesc == "" {
			errorDesc = "Authorization failed"
		}
		httpx.WriteJSONError(w, "auth_failed", errorDesc, http.StatusUnauthorized)
		return
	}

	tokens, err := h.exchangeCode(code)
	if err != nil {
		httpx.WriteJSONError(w, "token_exchange_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	userInfo, err := h.railway.FetchUserInfo(tokens.AccessToken)
	if err != nil {
		httpx.WriteJSONError(w, "user_info_failed", "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	hasAccess, err := h.railway.UserHasProjectAccess(tokens.AccessToken, h.cfg.RailwayProjectID)
	if err != nil {
		httpx.WriteJSONError(w, "project_check_failed", "Failed to check project access", http.StatusInternalServerError)
		return
	}

	if !hasAccess {
		loginErrURL := h.cfg.URI(config.RouteLogin, config.PathOnly) + "?error=no_access"
		http.Redirect(w, r, loginErrURL, http.StatusTemporaryRedirect)
		return
	}

	sess := h.session.CreateSession(userInfo.Sub, userInfo.Email, userInfo.Name, tokens.AccessToken)
	if err := h.session.SetSessionCookie(w, r, sess); err != nil {
		httpx.WriteJSONError(w, "session_error", "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	h.session.ClearSessionCookie(w, r)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// handleAuthError renders a human-readable HTML error page. It is used when
// the login route is reached with an ?error= query parameter, indicating that
// a previous OAuth attempt completed but was rejected (e.g. wrong project).
func (h *Handler) handleAuthError(w http.ResponseWriter, r *http.Request, errType string) {
	var message string
	switch errType {
	case "no_access":
		message = "Your Railway account does not have access to this project. Make sure you granted the application access to the correct project during authorization."
	default:
		message = "An authentication error occurred. Please try again."
	}

	loginURL := h.cfg.URI(config.RouteLogin, config.PathOnly)
	staticRoot := h.cfg.AuthPrefix + "/static"

	h.renderer.RenderUnauthenticated(w, views.UnauthData{
		StaticRoot:   staticRoot,
		Message:      message,
		LoginURL:     loginURL,
		ReconsentURL: loginURL + "?reconsent=true",
	}, http.StatusForbidden)
}

func (h *Handler) exchangeCode(code string) (*tokenResponse, error) {
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {h.cfg.URI(config.RouteCallback, config.FullURL)},
	}

	req, err := http.NewRequest("POST", oauthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(h.cfg.RailwayClientID, h.cfg.RailwayClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokens tokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &tokens, nil
}

// isSafeRedirect returns true only for relative paths, preventing open redirects.
func isSafeRedirect(redirectURL string) bool {
	return strings.HasPrefix(redirectURL, "/") && !strings.HasPrefix(redirectURL, "//")
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
