package models

import (
	tokenmodels "github.com/horizoncd/horizon/pkg/token/models"
)

type AccessToken struct {
	tokenmodels.Token
	Role string
}
