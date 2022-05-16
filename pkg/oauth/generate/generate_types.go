package generate

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/oauth/models"
	"g.hz.netease.com/horizon/pkg/oauth/store"
	"golang.org/x/net/context"
)

type CodeGenerateInfo struct {
	Client  store.ClientInfo
	Token   models.Token // AccessToken
	Request *http.Request
}

type AuthorizationCodeGenerate interface {
	GenCode(ctx context.Context, info *CodeGenerateInfo) (code string, _ error)
}

type AccessTokenCodeGenerate interface {
	GetCode(ctx context.Context, info *CodeGenerateInfo) (accessCode string, _ error)
}
