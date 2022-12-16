package generate

import (
	"net/http"

	"github.com/horizoncd/horizon/pkg/oauth/models"
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
