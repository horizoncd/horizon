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
	"time"
)

// type userType uint8

// const (
// 	user  userType = 1
// 	robot userType = 2
// )

type OwnerType uint8

const (
	GroupOwnerType OwnerType = 1
)

type AppType uint8

const (
	HorizonOAuthAPP AppType = 1
	DirectOAuthAPP  AppType = 2
)

type OauthApp struct {
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"column:name"`
	ClientID    string    `gorm:"column:client_id"`
	RedirectURL string    `gorm:"column:redirect_url"`
	HomeURL     string    `gorm:"column:home_url"`
	Desc        string    `gorm:"column:description"`
	OwnerType   OwnerType `gorm:"column:owner_type"`
	OwnerID     uint      `gorm:"column:owner_id"`
	AppType     AppType   `gorm:"column:app_type"`

	CreatedAt time.Time `gorm:"column:created_at"`
	CreatedBy uint      `gorm:"column:created_by"`

	UpdatedAt time.Time `gorm:"column:updated_at"`
	UpdatedBy uint      `gorm:"column:updated_by"`
}

func (a *OauthApp) IsGroupOwnerType() bool {
	return a.OwnerType == GroupOwnerType
}

type OauthClientSecret struct {
	ID           uint      `gorm:"column:id" json:"id"`
	ClientID     string    `gorm:"column:client_id" json:"clientID"`
	ClientSecret string    `gorm:"column:client_secret" json:"clientSecret"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
	CreatedBy    uint      `gorm:"column:created_by" json:"createdBy"`
}
