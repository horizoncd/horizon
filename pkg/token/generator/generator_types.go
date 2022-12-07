package generator

import (
	"net/http"

	"g.hz.netease.com/horizon/pkg/token/models"
)

type CodeGenerateInfo struct {
	Token   models.Token
	Request *http.Request
}

type AuthorizationCodeGenerator interface {
	GenCode(info *CodeGenerateInfo) string
}

type AccessTokenCodeGenerator interface {
	GenCode(info *CodeGenerateInfo) string
}
