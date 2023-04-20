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

package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
)

const (
	UserTypeCommon = iota
	UserTypeRobot
)

type User struct {
	global.Model

	Name     string
	FullName string
	Email    string
	Phone    string
	UserType uint
	Password string
	OidcID   string `gorm:"column:oidc_id"`
	OidcType string `gorm:"column:oidc_type"`
	Admin    bool
	Banned   bool
}

type UserBasic struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func ToUser(user *User) *UserBasic {
	if user == nil {
		return nil
	}
	return &UserBasic{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}
