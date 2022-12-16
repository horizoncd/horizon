package models

import (
	oauthmodels "github.com/horizoncd/horizon/pkg/oauth/models"
)

type AccessToken struct {
	oauthmodels.Token
	Role string
}
