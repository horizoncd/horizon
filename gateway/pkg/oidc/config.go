package oidc

// Config oidc config including clientID,
// clientSecret, endpoint, redirectURI and scopes
type Config struct {
	ClientID     string
	ClientSecret string
	Endpoint     Endpoint
	RedirectURI  string
	Scopes       []string
}

// Endpoint oidc endpoint including authURL, tokenURL, userURL:
// 1. authURL used for authentication, which will return a code when it calls back
// 2. tokenURL used for getting accessToken with code
// 3. userURL used for getting userinfo with accessToken
type Endpoint struct {
	AuthURL  string
	TokenURL string
	UserURL  string
}

// User userinfo got from oidc
type User struct {
	ID    string
	Name  string
	Email string
	Phone string
}
