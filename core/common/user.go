package common

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	herror "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/authentication/user"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

const (
	UserQueryName = "filter"
	UserQueryType = "userType"
)

const (
	contextUserKey = "contextUser"

	AuthorizationHeaderKey = "Authorization"
	TokenHeaderValuePrefix = "Bearer"
)

func UserContextKey() string {
	return contextUserKey
}

func UserFromContext(ctx context.Context) (user.User, error) {
	u, ok := ctx.Value(contextUserKey).(user.User)
	if !ok {
		return nil, herror.ErrFailedToGetUser
	}
	return u, nil
}

func WithContext(parent context.Context, user user.User) context.Context {
	return context.WithValue(parent, UserContextKey(), user) //nolint
}

func SetUser(c *gin.Context, user user.User) {
	// attach user to context
	c.Set(contextUserKey, user)
}

func GetToken(c *gin.Context) (string, error) {
	if _, ok := c.Request.Header[AuthorizationHeaderKey]; !ok {
		return "", perror.Wrap(herror.ErrAuthorizationHeaderNotFound, "")
	}
	token, err := func() (string, error) {
		parts := strings.Split(c.Request.Header.Get(AuthorizationHeaderKey), " ")
		if len(parts) != 2 || parts[0] != TokenHeaderValuePrefix {
			return "", perror.Wrapf(herror.ErrOAuthTokenFormatError, "header = %s",
				c.Request.Header.Get(AuthorizationHeaderKey))
		}
		return parts[1], nil
	}()
	return token, err
}
