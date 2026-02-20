package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"turnstile/internal/httpx"
)

const (
	sessionCookieName = "railway_session"
	sessionDuration   = 1 * time.Hour
)

type Session struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Manager struct {
	secretKey []byte
}

func NewManager(secret string) (*Manager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("secret must be at least 32 characters")
	}
	key := []byte(secret[:32])
	return &Manager{secretKey: key}, nil
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
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	encrypted, err := sm.encrypt(data)
	if err != nil {
		return fmt.Errorf("encrypt session: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(encrypted)

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
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

	decoded, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("decode cookie: %w", err)
	}

	decrypted, err := sm.decrypt(decoded)
	if err != nil {
		return nil, fmt.Errorf("decrypt session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(decrypted, &session); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, nil
	}

	return &session, nil
}

func (sm *Manager) ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
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

func (sm *Manager) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sm.secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func (sm *Manager) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sm.secretKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
