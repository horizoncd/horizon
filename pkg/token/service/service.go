package service

import (
	"context"
	"strings"
	"time"

	herror "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"g.hz.netease.com/horizon/pkg/token/generator"
	tokenmanager "g.hz.netease.com/horizon/pkg/token/manager"
	tokenmodels "g.hz.netease.com/horizon/pkg/token/models"
)

type Service interface {
	// CreateUserAccessToken used for personal access Token and resource access Token
	CreateUserAccessToken(ctx context.Context, name, expiresAtStr string,
		userID uint, scopes []string) (*tokenmodels.Token, error)
	// CreateInternalAccessToken used for internal component to access horizon api
	CreateInternalAccessToken(ctx context.Context, name string, expiresIn time.Duration,
		userID uint, scopes []string) (*tokenmodels.Token, error)
}

func NewService(manager *managerparam.Manager) Service {
	return &service{
		tokenManager:             manager.TokenManager,
		userAccessTokenGenerator: generator.NewUserAccessTokenGenerator(),
	}
}

type service struct {
	tokenManager             tokenmanager.Manager
	userAccessTokenGenerator generator.AccessTokenCodeGenerator
}

func (s *service) CreateUserAccessToken(ctx context.Context, name, expiresAtStr string,
	userID uint, scopes []string) (*tokenmodels.Token, error) {
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
	token := s.genUserAccessToken(name, userID, scopes, createdAt, expiresIn)
	// 3. create token in db
	token, err := s.tokenManager.CreateToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *service) CreateInternalAccessToken(ctx context.Context, name string, expiresIn time.Duration,
	userID uint, scopes []string) (*tokenmodels.Token, error) {
	// 1. check expiration duration
	if expiresIn < 0 {
		return nil, perror.Wrap(herror.ErrParamInvalid, "expiration duration must be greater than 0")
	}
	// 2. generate user access token
	createdAt := time.Now()
	token := s.genUserAccessToken(name, userID, scopes, createdAt, expiresIn)
	// 3. create token in db
	token, err := s.tokenManager.CreateToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *service) genUserAccessToken(name string, userID uint, scopes []string,
	createdAt time.Time, expiresIn time.Duration) *tokenmodels.Token {
	return &tokenmodels.Token{
		Name: name,
		Code: s.userAccessTokenGenerator.GenCode(&generator.CodeGenerateInfo{
			Token: tokenmodels.Token{UserID: userID},
		}),
		Scope:     strings.Join(scopes, " "),
		CreatedAt: createdAt,
		ExpiresIn: expiresIn,
		UserID:    userID,
	}
}
