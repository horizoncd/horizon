package models

import (
	"database/sql/driver"
	"fmt"

	"g.hz.netease.com/horizon/pkg/server/global"
)

type IdentityProvider struct {
	global.Model

	DisplayName             string
	Name                    string
	Avatar                  string
	AuthorizationEndpoint   string
	TokenEndpoint           string
	UserinfoEndpoint        string
	RevocationEndpoint      string
	Issuer                  string
	Scopes                  string
	SigningAlgs             string
	TokenEndpointAuthMethod *TokenEndpointAuthMethod
	Jwks                    string
	ClientID                string
	ClientSecret            string
}

type TokenEndpointAuthMethod uint8

const (
	ClientSecretSentAsPost = iota + 1
	ClientSecretSentAsBasicAuth
	ClientSecretAsJwt
	JwtSignedWithPrivateKey
)

const (
	ClientSecretSentAsPostStr      = "client_secret_sent_as_post"
	ClientSecretSentAsBasicAuthStr = "client_secret_sent_as_basic_auth"
	ClientSecretAsJwtStr           = "client_secret_as_jwt"
	JwtSignedWithPrivateKeyStr     = "jwt_signed_with_private_key"
)

func (t *TokenEndpointAuthMethod) String() (string, error) {
	switch *t {
	case ClientSecretSentAsPost:
		return ClientSecretSentAsPostStr, nil
	case ClientSecretSentAsBasicAuth:
		return ClientSecretSentAsBasicAuthStr, nil
	case ClientSecretAsJwt:
		return ClientSecretAsJwtStr, nil
	case JwtSignedWithPrivateKey:
		return JwtSignedWithPrivateKeyStr, nil
	}
	return "", fmt.Errorf("unsupported value: %v", t)
}

func (t *TokenEndpointAuthMethod) Scan(value interface{}) error {
	bts, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal TokenEndpointAuthMethod from value: %v", value)
	}
	str := string(bts)
	switch str {
	case ClientSecretSentAsPostStr:
		*t = ClientSecretSentAsPost
		return nil
	case ClientSecretSentAsBasicAuthStr:
		*t = ClientSecretSentAsBasicAuth
		return nil
	case ClientSecretAsJwtStr:
		*t = ClientSecretAsJwt
		return nil
	case JwtSignedWithPrivateKeyStr:
		*t = JwtSignedWithPrivateKey
		return nil
	}
	return fmt.Errorf("failed to unmarshal TokenEndpointAuthMethod: unsupported value %v", str)
}

func (t *TokenEndpointAuthMethod) Value() (driver.Value, error) {
	return t.String()
}

func (t *TokenEndpointAuthMethod) MarshalJSON() ([]byte, error) {
	str, err := t.String()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

func (t *TokenEndpointAuthMethod) UnmarshalJSON(data []byte) error {
	str := string(data)
	switch str {
	case fmt.Sprintf("\"%s\"", ClientSecretSentAsPostStr):
		*t = ClientSecretSentAsPost
		return nil
	case fmt.Sprintf("\"%s\"", ClientSecretSentAsBasicAuthStr):
		*t = ClientSecretSentAsBasicAuth
		return nil
	case fmt.Sprintf("\"%s\"", ClientSecretAsJwtStr):
		*t = ClientSecretAsJwt
		return nil
	case fmt.Sprintf("\"%s\"", JwtSignedWithPrivateKeyStr):
		*t = JwtSignedWithPrivateKey
		return nil
	}
	return fmt.Errorf("failed to unmarshal TokenEndpointAuthMethod: unsupported value %v", str)
}
