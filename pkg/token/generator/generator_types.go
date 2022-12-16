package generator

import (
	"net/http"

	"github.com/horizoncd/horizon/pkg/token/models"
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
