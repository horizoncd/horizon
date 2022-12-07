package generator

import (
	"bytes"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ref: https://github.blog/2021-04-05-behind-githubs-new-authentication-token-formats/
const (
	HorizonAppUserToServerAccessTokenPrefix = "hu_"
	OauthAPPAccessTokenPrefix               = "ho_"
	AccessTokenPrefix                       = "ha_"
	// InternalAccessTokenPrefix for internal component access horizon api, such as tekton
	InternalAccessTokenPrefix = "hi_"
	// HorizonAppServerToServerAccessTokenPrefix = "hs_"
)

func NewAuthorizeGenerator() AuthorizationCodeGenerator {
	return &AuthorizeGenerator{}
}

func NewHorizonAppUserToServerAccessGenerator() AccessTokenCodeGenerator {
	return &BasicAccessTokenGenerator{prefix: HorizonAppUserToServerAccessTokenPrefix}
}

func NewOauthAccessGenerator() AccessTokenCodeGenerator {
	return &BasicAccessTokenGenerator{prefix: OauthAPPAccessTokenPrefix}
}

func NewUserAccessTokenGenerator() AccessTokenCodeGenerator {
	return &UserAccessTokenGenerator{prefix: AccessTokenPrefix}
}

func NewInternalAccessTokenGenerator() AccessTokenCodeGenerator {
	return &UserAccessTokenGenerator{prefix: InternalAccessTokenPrefix}
}

type AuthorizeGenerator struct{}

type BasicAccessTokenGenerator struct {
	prefix string
}

type UserAccessTokenGenerator struct {
	prefix string
}

func (g *AuthorizeGenerator) GenCode(info *CodeGenerateInfo) string {
	buf := bytes.NewBufferString(info.Token.ClientID)
	buf.WriteString(strconv.Itoa(int(info.Token.UserID)))
	token := uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes())
	code := base64.URLEncoding.EncodeToString([]byte(token.String()))
	code = strings.ToUpper(strings.TrimRight(code, "="))
	return code
}

func (g *BasicAccessTokenGenerator) GenCode(info *CodeGenerateInfo) string {
	buf := bytes.NewBufferString(info.Token.ClientID)
	buf.WriteString(strconv.Itoa(int(info.Token.UserID)))
	buf.WriteString(strconv.FormatInt(info.Token.CreatedAt.UnixNano(), 10))
	access := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	access = g.prefix + strings.ToUpper(strings.TrimRight(access, "="))
	return access
}

func (g *UserAccessTokenGenerator) GenCode(info *CodeGenerateInfo) string {
	buf := bytes.NewBufferString(time.Now().String())
	buf.WriteString(strconv.Itoa(int(info.Token.UserID)))
	code := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	code = g.prefix + strings.ToUpper(strings.TrimRight(code, "="))
	return code
}
