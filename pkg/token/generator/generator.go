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
	"k8s.io/apimachinery/pkg/util/rand"
)

// ref: https://github.blog/2021-04-05-behind-githubs-new-authentication-token-formats/
const (
	HorizonAppUserToServerAccessTokenPrefix = "hu_"
	OauthAPPAccessTokenPrefix               = "ho_"
	AccessTokenPrefix                       = "ha_"
	RefreshTokenPrefix                      = "hr_"
)

func NewAuthorizeGenerator() CodeGenerator {
	return &authorizationCodeGenerator{}
}

func NewHorizonAppUserToServerAccessGenerator() CodeGenerator {
	return &basicTokenGenerator{prefix: HorizonAppUserToServerAccessTokenPrefix}
}

func NewOauthAccessGenerator() CodeGenerator {
	return &basicTokenGenerator{prefix: OauthAPPAccessTokenPrefix}
}

func NewGeneralAccessTokenGenerator() CodeGenerator {
	return &basicTokenGenerator{prefix: AccessTokenPrefix}
}

func NewRefreshTokenGenerator() CodeGenerator {
	return &basicTokenGenerator{prefix: RefreshTokenPrefix}
}

type authorizationCodeGenerator struct{}

type basicTokenGenerator struct {
	prefix string
}

func (g *authorizationCodeGenerator) Generate(info *CodeGenerateInfo) string {
	buf := bytes.NewBufferString(info.Token.ClientID)
	buf.WriteString(strconv.Itoa(int(info.Token.UserID)))
	token := uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes())
	code := base64.URLEncoding.EncodeToString([]byte(token.String()))
	code = strings.ToUpper(strings.TrimRight(code, "="))
	return code
}

func (g *basicTokenGenerator) Generate(info *CodeGenerateInfo) string {
	clientID := func(info *CodeGenerateInfo) string {
		if info.Token.ClientID != "" {
			return info.Token.ClientID
		}
		return rand.String(20)
	}(info)
	buf := bytes.NewBufferString(clientID)
	buf.WriteString(strconv.Itoa(int(info.Token.UserID)))
	buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
	access := base64.URLEncoding.EncodeToString([]byte(uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes()).String()))
	return g.prefix + strings.ToUpper(strings.TrimRight(access, "="))
}
