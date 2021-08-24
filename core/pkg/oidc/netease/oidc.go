package netease

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"g.hz.netease.com/horizon/core/pkg/oidc"
)

type OIDC struct {
	config *oidc.Config
}

func NewOIDC(config *oidc.Config) *OIDC {
	return &OIDC{
		config: config,
	}
}

type User struct {
	HzNumber string `json:"hzNumber"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

func (o *OIDC) GetRedirectURL(requestHost, state string) string {
	params := url.Values{
		"response_type": {"code"},
		"scope":         {strings.Join(o.config.Scopes, " ")},
		"client_id":     {o.config.ClientID},
		"state":         {state},
		"redirect_uri":  {httpScheme + "://" + requestHost + o.config.RedirectURI},
	}
	return fmt.Sprintf("%s?%s", o.config.Endpoint.AuthURL, params.Encode())
}

const httpScheme = "http"

type tokenStruct struct {
	AccessToken string `json:"access_token"`
}

func (o *OIDC) getAccessToken(requestHost, code string)(string, error) {
	redirectURL := url.QueryEscape(httpScheme + "://" + requestHost + o.config.RedirectURI)

	req, err := http.NewRequest(http.MethodPost, o.config.Endpoint.TokenURL, strings.NewReader(url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURL},
		"client_id":     {o.config.ClientID},
		"client_secret": {o.config.ClientSecret},
	}.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("authorization_code from oidc error with code: %d, body: %v",
			resp.StatusCode, string(body))
	}
	var ts *tokenStruct
	if err := json.Unmarshal(body, &ts); err != nil {
		return "", err
	}
	return ts.AccessToken, nil
}

func (o *OIDC) GetUser(requestHost, code string) (*oidc.User, error) {
	accessToken, err := o.getAccessToken(requestHost, code)
	if err != nil {
		return nil, err
	}

	userInfoUrl := fmt.Sprintf("%s?%s", o.config.Endpoint.UserURL, url.Values{
		"access_token": {accessToken},
	}.Encode())
	resp, err := http.Get(userInfoUrl)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get userinfo from oidc error with code: %d, body: %v",
			resp.StatusCode, string(body))
	}

	var user *User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}
	return &oidc.User{
		ID:    user.HzNumber,
		Name:  user.FullName,
		Email: user.Email,
	}, nil
}
