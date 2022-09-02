package generate

import (
	"bytes"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func NewAuthorizeGenerate() *AuthorizeGenerate {
	return &AuthorizeGenerate{}
}

type AuthorizeGenerate struct{}

func (g *AuthorizeGenerate) GenCode(info *CodeGenerateInfo) (code string) {
	buf := bytes.NewBufferString(info.Token.ClientID)
	buf.WriteString(info.Token.UserOrRobotIdentity)
	token := uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes())
	code = base64.URLEncoding.EncodeToString([]byte(token.String()))
	code = strings.ToUpper(strings.TrimRight(code, "="))
	return code
}

// ref: https://github.blog/2021-04-05-behind-githubs-new-authentication-token-formats/
const (
	HorizonAppUserToServerAccessTokenPrefix = "hu_"
	OauthAPPAccessTokenPrefix               = "ho_"
	// HorizonAppServerToServerAccessTokenPrefix = "hs_"
)

func NewOauthAccessGenerate() *BasicAccessTokenGenerate {
	return &BasicAccessTokenGenerate{prefix: OauthAPPAccessTokenPrefix}
}

func NewHorizonAppUserToServerAccessGenerate() *BasicAccessTokenGenerate {
	return &BasicAccessTokenGenerate{prefix: HorizonAppUserToServerAccessTokenPrefix}
}

type BasicAccessTokenGenerate struct {
	prefix string
}

func (g *BasicAccessTokenGenerate) GenCode(info *CodeGenerateInfo) (code string) {
	buf := bytes.NewBufferString(info.Token.ClientID)
	buf.WriteString(info.Token.UserOrRobotIdentity)
	buf.WriteString(strconv.FormatInt(info.Token.CreatedAt.UnixNano(), 10))
	access := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	access = g.prefix + strings.ToUpper(strings.TrimRight(access, "="))
	return access
}
