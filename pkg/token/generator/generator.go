// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	RefreshTokenPrefix                      = "hr_"
)

func NewAuthorizeGenerator() CodeGenerator {
	return &AuthorizeGenerator{}
}

func NewHorizonAppUserToServerAccessGenerator() CodeGenerator {
	return &BasicAccessTokenGenerator{prefix: HorizonAppUserToServerAccessTokenPrefix}
}

func NewOauthAccessGenerator() CodeGenerator {
	return &BasicAccessTokenGenerator{prefix: OauthAPPAccessTokenPrefix}
}

func NewGeneralAccessTokenGenerator() CodeGenerator {
	return &GeneralAccessTokenGenerator{prefix: AccessTokenPrefix}
}

func NewRefreshTokenGenerator() CodeGenerator {
	return &BasicRefreshTokenGenerator{prefix: RefreshTokenPrefix}
}

type AuthorizeGenerator struct{}

type BasicAccessTokenGenerator struct {
	prefix string
}

type GeneralAccessTokenGenerator struct {
	prefix string
}

type BasicRefreshTokenGenerator struct {
	prefix string
}

func (g *AuthorizeGenerator) Generate(info *CodeGenerateInfo) string {
	buf := bytes.NewBufferString(info.Token.ClientID)
	buf.WriteString(strconv.Itoa(int(info.Token.UserID)))
	token := uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes())
	code := base64.URLEncoding.EncodeToString([]byte(token.String()))
	code = strings.ToUpper(strings.TrimRight(code, "="))
	return code
}

func (g *BasicAccessTokenGenerator) Generate(info *CodeGenerateInfo) string {
	access := encodeInfoWithClient(info)
	code := g.prefix + strings.ToUpper(strings.TrimRight(access, "="))
	return code
}

func (g *GeneralAccessTokenGenerator) Generate(info *CodeGenerateInfo) string {
	buf := bytes.NewBufferString(time.Now().String())
	buf.WriteString(strconv.Itoa(int(info.Token.UserID)))
	code := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	code = g.prefix + strings.ToUpper(strings.TrimRight(code, "="))
	return code
}

func (g *BasicRefreshTokenGenerator) Generate(info *CodeGenerateInfo) string {
	access := encodeInfoWithClient(info)
	code := g.prefix + strings.ToUpper(strings.TrimRight(access, "="))
	return code
}

func encodeInfoWithClient(info *CodeGenerateInfo) string {
	buf := bytes.NewBufferString(info.Token.ClientID)
	buf.WriteString(strconv.Itoa(int(info.Token.UserID)))
	buf.WriteString(strconv.FormatInt(info.Token.CreatedAt.UnixNano(), 10))
	access := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	return access
}
