package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"turnstile/internal/httpx"
)

const (
	sessionCookieName = "railway_session"
	sessionDuration   = 1 * time.Hour
)

type Session struct {
	UserID      string
	Email       string
	Name        string
	AccessToken string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

func (sm *Manager) CreateSession(userID, email, name, accessToken string) *Session {
	now := time.Now()
	return &Session{
		UserID:      userID,
		Email:       email,
		Name:        name,
		AccessToken: accessToken,
		ExpiresAt:   now.Add(sessionDuration),
		CreatedAt:   now,
	}
}

func (sm *Manager) SetSessionCookie(w http.ResponseWriter, r *http.Request, session *Session) error {
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("generate session token: %w", err)
	}

	sm.mu.Lock()
	sm.sessions[token] = session
	sm.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   httpx.IsHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (sm *Manager) GetSession(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, nil
		}
		return nil, fmt.Errorf("get cookie: %w", err)
	}

	sm.mu.RLock()
	session, ok := sm.sessions[cookie.Value]
	sm.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	if time.Now().After(session.ExpiresAt) {
		sm.mu.Lock()
		delete(sm.sessions, cookie.Value)
		sm.mu.Unlock()
		return nil, nil
	}

	return session, nil
}

func (sm *Manager) ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		sm.mu.Lock()
		delete(sm.sessions, cookie.Value)
		sm.mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   httpx.IsHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
