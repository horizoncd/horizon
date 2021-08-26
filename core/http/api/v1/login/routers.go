package login

import (
	"net/http"
	"os"
	"regexp"
	"time"

	sessionmiddleware "g.hz.netease.com/horizon/core/middleware/session"
	"g.hz.netease.com/horizon/core/pkg/oidc"
	"g.hz.netease.com/horizon/core/pkg/oidc/netease"
	"g.hz.netease.com/horizon/core/pkg/session"
	"g.hz.netease.com/horizon/pkg/redis"
	"g.hz.netease.com/horizon/server/middleware"
	"g.hz.netease.com/horizon/server/middleware/log"
	"g.hz.netease.com/horizon/server/middleware/requestid"
	"g.hz.netease.com/horizon/server/route"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/api/v1")
	var c, _ = NewController()
	api.Use(requestid.Middleware())
	api.Use(log.Middleware())
	api.Use(sessionmiddleware.Middleware(c.sessionManager,
		middleware.MethodAndPathSkipper(http.MethodPost, regexp.MustCompile("^/api/v1/login")),
		middleware.MethodAndPathSkipper(http.MethodGet, regexp.MustCompile("^/api/v1/login/callback")),
	))

	var routes = route.Routes{
		{
			"Login",
			http.MethodPost,
			"/login",
			c.Login,
		}, {
			"LoginCallback",
			http.MethodGet,
			"/login/callback",
			c.Callback,
		}, {
			"Logout",
			http.MethodPost,
			"/logout",
			c.Logout,
		}, {
			"LoginStatus",
			http.MethodGet,
			"/login/status",
			c.UserStatus,
		},
	}
	route.RegisterRoutes(api, routes)
}

type Controller struct {
	oidc           oidc.Interface
	sessionManager session.Interface
}

func NewController() (*Controller, error) {
	redisPool, err := redis.GetRedisPool("core", os.Getenv("REDIS_URL"), nil)
	redisHelper, err := redis.NewHelperWithPool(redisPool,
		redis.NewOptionsWithDefaultCodec(
			redis.Prefix("core:session:"), redis.Expiration(2*time.Hour)))
	if err != nil {
		return nil, err
	}

	return &Controller{
		oidc: netease.NewOIDC(&oidc.Config{
			ClientID:     "",
			ClientSecret: "",
			Endpoint: oidc.Endpoint{
				AuthURL:  "https://oidc.mockserver.org/connect/authorize",
				TokenURL: "https://oidc.mockserver.org/connect/token",
				UserURL:  "https://oidc.mockserver.org/connect/userinfo",
			},
			RedirectURI: "/api/v1/login/callback",
			Scopes:      []string{"openid", "fullname", "nickname", "email"},
		}),
		sessionManager: session.NewRedisSession(redisHelper),
	}, nil
}
