package oauth

import "time"

type TokenResponse struct {
	AccessToken string
	ExpireIn    time.Duration
	Scope       string
	TokenType   string
}

type Controller interface {
	GenAuthorizeCode(clientID string, redirectURL string, state string) (code string, _ error)
	GetAuthorizeCode(clientID string, redirectURL string, state string) (code string, _ error)
	GetAccessCode(clientID, clientSecret, code, redirectURL, state string) (token TokenResponse, _ error)
}
