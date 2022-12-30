package token

import "time"

type Config struct {
	// JwtSigningKey is used to sign JWT tokens
	JwtSigningKey string `yaml:"jwtSigningKey"`
	// CallbackTokenExpireIn is the expiration time of token for tekton callback
	CallbackTokenExpireIn time.Duration `yaml:"callbackTokenExpireIn"`
}
