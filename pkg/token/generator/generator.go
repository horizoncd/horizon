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
	// HorizonAppServerToServerAccessTokenPrefix = "hs_"
)

func NewAuthorizeGenerator() *AuthorizeGenerator {
	return &AuthorizeGenerator{}
}

func NewHorizonAppUserToServerAccessGenerator() *BasicAccessTokenGenerator {
	return &BasicAccessTokenGenerator{prefix: HorizonAppUserToServerAccessTokenPrefix}
}

func NewOauthAccessGenerator() *BasicAccessTokenGenerator {
	return &BasicAccessTokenGenerator{prefix: OauthAPPAccessTokenPrefix}
}

func NewUserAccessTokenGenerator() *UserAccessTokenGenerator {
	return &UserAccessTokenGenerator{prefix: AccessTokenPrefix}
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
