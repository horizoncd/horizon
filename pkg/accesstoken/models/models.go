package models

import (
	oauthmodels "g.hz.netease.com/horizon/pkg/oauth/models"
)

type AccessToken struct {
	oauthmodels.Token
	Role string
}
