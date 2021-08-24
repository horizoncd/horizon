package session

import "g.hz.netease.com/horizon/core/pkg/oidc"

type Interface interface {
	GetSession(sessionID string) (*Session, error)
	SetSession(sessionID string, session *Session) error
	DeleteSession(sessionID string) error
}

type Session struct {
	FromHost    string
	RedirectURL string
	User        *oidc.User
}