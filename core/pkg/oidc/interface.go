package oidc

import "context"

type Interface interface {
	// GetRedirectURL 获取需要重定向到的oidc认证页面
	GetRedirectURL(ctx context.Context, requestHost, state string) string
	// GetUser 根据code获取User
	GetUser(ctx context.Context, requestHost, code string) (*User, error)
}
