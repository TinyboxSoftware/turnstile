package auth

import (
	"context"

	"railway-oauth-proxy/internal/session"
)

type contextKey string

const sessionContextKey contextKey = "session"

func SetSessionContext(ctx context.Context, sess *session.Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, sess)
}

func GetSessionFromContext(ctx context.Context) *session.Session {
	if sess, ok := ctx.Value(sessionContextKey).(*session.Session); ok {
		return sess
	}
	return nil
}
