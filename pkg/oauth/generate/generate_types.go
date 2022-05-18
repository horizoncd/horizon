package generate

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/oauth/models"
)

type CodeGenerateInfo struct {
	Token   models.Token // AccessToken
	Request *http.Request
}

type AuthorizationCodeGenerate interface {
	GenCode(info *CodeGenerateInfo) (code string)
}

type AccessTokenCodeGenerate interface {
	GenCode(info *CodeGenerateInfo) (accessCode string)
}
