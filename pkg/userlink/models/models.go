package models

import "github.com/horizoncd/horizon/pkg/server/global"

type UserLink struct {
	global.Model

	Sub       string
	IdpID     uint
	UserID    uint
	Name      string
	Email     string
	Deletable bool
}

func (UserLink) TableName() string {
	return "tb_idp_user"
}
