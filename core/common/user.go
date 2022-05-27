package common

import (
	"context"
	"errors"
	"strings"

	herror "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/authentication/user"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserKey         = "contextUser"
	AuthorizationHeaderKey = "Authorization"
	TokenHeaderValuePrefix = "token"
)

func Key() string {
	return ContextUserKey
}

func FromContext(ctx context.Context) (user.User, error) {
	u, ok := ctx.Value(ContextUserKey).(user.User)
	if !ok {
		return nil, herror.ErrFailedToGetUser
	}
	return u, nil
}

func WithContext(parent context.Context, user user.User) context.Context {
	return context.WithValue(parent, Key(), user) // nolint
}

func SetUser(c *gin.Context, user user.User) {
	// attach user to context
	c.Set(ContextUserKey, user)
}

func GetToken(c *gin.Context) (string, error) {
	if _, ok := c.Request.Header[AuthorizationHeaderKey]; !ok {
		return "", errors.New("header not found " + AuthorizationHeaderKey)
	}
	token, err := func() (string, error) {
		parts := strings.Split(c.Request.Header.Get(AuthorizationHeaderKey), " ")
		if len(parts) != 2 && parts[0] == TokenHeaderValuePrefix {
			return "", errors.New("token format error")
		}
		return parts[1], nil
	}()
	return token, err
}
