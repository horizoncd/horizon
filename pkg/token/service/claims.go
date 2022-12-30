package service

import "github.com/golang-jwt/jwt/v4"

const ClaimsIssuer = "horizon"

type Claims struct {
	UserID        uint
	PipelinerunID *uint
	jwt.RegisteredClaims
}

type ClaimsOption func(*Claims)

func WithUserID(userID uint) ClaimsOption {
	return func(claims *Claims) {
		claims.UserID = userID
	}
}

func WithUserIDAndPipelinerunID(userID, pipelinerunID uint) ClaimsOption {
	return func(claims *Claims) {
		claims.UserID = userID
		claims.PipelinerunID = &pipelinerunID
	}
}
