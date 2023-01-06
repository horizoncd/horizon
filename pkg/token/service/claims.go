package service

import "github.com/golang-jwt/jwt/v4"

const ClaimsIssuer = "horizon"

type Claims struct {
	PipelinerunID *uint
	jwt.RegisteredClaims
}

type ClaimsOption func(*Claims)

func WithPipelinerunID(pipelinerunID uint) ClaimsOption {
	return func(claims *Claims) {
		claims.PipelinerunID = &pipelinerunID
	}
}
