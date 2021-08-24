package oidc

type Interface interface{
	// GetRedirectURL 获取需要重定向到的oidc认证页面
	GetRedirectURL(requestHost, state string) string
	// GetUser 根据code获取User
	GetUser(requestHost, code string) (*User, error)
}
