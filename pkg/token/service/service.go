package service

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	herror "github.com/horizoncd/horizon/core/errors"
	tokenconfig "github.com/horizoncd/horizon/pkg/config/token"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/token/generator"
	tokenmanager "github.com/horizoncd/horizon/pkg/token/manager"
	tokenmodels "github.com/horizoncd/horizon/pkg/token/models"
)

type Service interface {
	// CreateAccessToken used for personal access Token and resource access Token
	CreateAccessToken(ctx context.Context, name, expiresAtStr string,
		userID uint, scopes []string) (*tokenmodels.Token, error)
	CreateJWTToken(subject string, expiresIn time.Duration, options ...ClaimsOption) (string, error)
	ParseJWTToken(tokenStr string) (Claims, error)
}

func NewService(manager *managerparam.Manager, config tokenconfig.Config) Service {
	return &service{
		tokenManager: manager.TokenManager,
		TokenConfig:  config,
	}
}

type service struct {
	tokenManager tokenmanager.Manager
	TokenConfig  tokenconfig.Config
}

func (s *service) CreateAccessToken(ctx context.Context, name, expiresAtStr string,
	userID uint, scopes []string,
) (*tokenmodels.Token, error) {
	// 1. check expiration date
	createdAt := time.Now()
	expiresIn := time.Duration(0)
	if expiresAtStr != NeverExpire {
		expiredAt, err := time.Parse(ExpiresAtFormat, expiresAtStr)
		if err != nil {
			return nil, perror.Wrapf(herror.ErrParamInvalid, "invalid expiration time, error: %s", err.Error())
		}
		if !expiredAt.After(createdAt) {
			return nil, perror.Wrap(herror.ErrParamInvalid, "expiration time must be later than current time")
		}
		expiresIn = expiredAt.Sub(createdAt)
	}
	// 2. generate user access token
	gen := generator.NewGeneralAccessTokenGenerator()
	token, err := s.genAccessToken(gen, name, userID, scopes, createdAt, expiresIn)
	if err != nil {
		return nil, err
	}
	// 3. create token in db
	token, err = s.tokenManager.CreateToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *service) genAccessToken(gen generator.AccessTokenCodeGenerator, name string, userID uint,
	scopes []string, createdAt time.Time, expiresIn time.Duration,
) (*tokenmodels.Token, error) {
	code := gen.GenCode(&generator.CodeGenerateInfo{
		Token: tokenmodels.Token{UserID: userID},
	})
	return &tokenmodels.Token{
		Name:      name,
		Code:      code,
		Scope:     strings.Join(scopes, " "),
		CreatedAt: createdAt,
		ExpiresIn: expiresIn,
		UserID:    userID,
	}, nil
}

func (s *service) CreateJWTToken(subject string, expiresIn time.Duration, options ...ClaimsOption) (string, error) {
	now := time.Now().UTC()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    ClaimsIssuer,
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	if expiresIn > 0 {
		expires := now.Add(expiresIn)
		claims.ExpiresAt = jwt.NewNumericDate(expires)
	}

	for _, opt := range options {
		opt(claims)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.TokenConfig.JwtSigningKey))
}

// ParseJWTToken parses string and return claims.
func (s *service) ParseJWTToken(tokenStr string) (Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, perror.Wrapf(herror.ErrTokenInvalid,
				"unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.TokenConfig.JwtSigningKey), nil
	})
	if err != nil {
		return Claims{}, err
	}

	if claims.Issuer != ClaimsIssuer {
		return Claims{}, perror.Wrapf(herror.ErrTokenInvalid,
			"unexpected claims issuer: %v", claims.Issuer)
	}
	return claims, nil
}
