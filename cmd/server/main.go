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
	authMiddleware := auth.NewMiddleware(sessionManager)

	proxyHandler, err := proxy.NewHandler(cfg.BackendURL)
	if err != nil {
		log.Fatalf("Failed to create proxy handler: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/oauth/login", oauthHandler.LoginHandler)
	mux.HandleFunc("/oauth/callback", oauthHandler.CallbackHandler)
	mux.HandleFunc("/oauth/logout", oauthHandler.LogoutHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("debug_request", "path", r.URL.Path, "query", r.URL.RawQuery, "cookies", r.Header.Get("Cookie"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"path": r.URL.Path, "query": r.URL.RawQuery})
	})

	mux.Handle("/", authMiddleware.RequireAuth(proxyHandler))

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("Starting server", "addr", addr)

	if err := http.ListenAndServe(addr, httpx.LoggingMiddleware(mux)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
