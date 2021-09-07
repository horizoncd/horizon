package session

import (
	"context"

	"g.hz.netease.com/horizon/gateway/pkg/oidc"
)

type Interface interface {
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	SetSession(ctx context.Context, sessionID string, session *Session) error
	DeleteSession(ctx context.Context, sessionID string) error
}

type Session struct {
	FromHost    string
	RedirectURL string
	User        *oidc.User
}