package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"turnstile/internal/auth"
	"turnstile/internal/config"
	"turnstile/internal/httpx"
	"turnstile/internal/oauth"
	"turnstile/internal/proxy"
	"turnstile/internal/railway"
	"turnstile/internal/session"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	sessionManager, err := session.NewManager(cfg.SessionSecret)
	if err != nil {
		log.Fatalf("Failed to create session manager: %v", err)
	}

	railwayClient := railway.NewClient(nil)
	oauthHandler := oauth.NewHandler(cfg, sessionManager, railwayClient)
	authMiddleware := auth.NewMiddleware(sessionManager, cfg.URI(config.RouteLogin, config.PathOnly))

	proxyHandler, err := proxy.NewHandler(cfg.BackendURL)
	if err != nil {
		log.Fatalf("Failed to create proxy handler: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc(cfg.URI(config.RouteLogin, config.PathOnly), oauthHandler.LoginHandler)
	mux.HandleFunc(cfg.URI(config.RouteLogout, config.PathOnly), oauthHandler.CallbackHandler)
	mux.HandleFunc(cfg.URI(config.RouteCallback, config.PathOnly), oauthHandler.LogoutHandler)

	mux.HandleFunc(cfg.URI(config.RouteHealth, config.PathOnly), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.Handle("/", authMiddleware.RequireAuth(proxyHandler))

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("Starting server", "addr", addr)

	if err := http.ListenAndServe(addr, httpx.LoggingMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
