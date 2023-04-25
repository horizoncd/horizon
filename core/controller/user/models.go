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

package user

import (
	"time"

	"github.com/horizoncd/horizon/pkg/user/models"
	lmodels "github.com/horizoncd/horizon/pkg/userlink/models"
)

type User struct {
	ID        uint      `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	FullName  string    `json:"fullName,omitempty"`
	Email     string    `json:"email,omitempty"`
	IsAdmin   bool      `json:"isAdmin"`
	IsBanned  bool      `json:"isBanned"`
	Phone     string    `json:"phone,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

func ofUser(u *models.User) *User {
	return &User{
		ID:        u.ID,
		Name:      u.Name,
		FullName:  u.FullName,
		Email:     u.Email,
		IsAdmin:   u.Admin,
		IsBanned:  u.Banned,
		Phone:     u.Phone,
		UpdatedAt: u.UpdatedAt,
		CreatedAt: u.CreatedAt,
	}
}

func ofUsers(users []*models.User) []*User {
	resp := make([]*User, 0, len(users))
	for _, u := range users {
		resp = append(resp, ofUser(u))
	}
	return resp
}

type UpdateUserRequest struct {
	IsAdmin  *bool `json:"isAdmin"`
	IsBanned *bool `json:"isBanned"`
}

type Link struct {
	ID         uint   `json:"id"`
	Sub        string `json:"sub"`
	IdpID      uint   `json:"idpId"`
	UserID     uint   `json:"userId"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Unlinkable bool   `json:"unlinkable"`
}

func ofUserLink(link *lmodels.UserLink) *Link {
	return &Link{
		ID:         link.ID,
		Sub:        link.Sub,
		IdpID:      link.IdpID,
		UserID:     link.UserID,
		Name:       link.Name,
		Email:      link.Email,
		Unlinkable: link.Deletable,
	}
}

func ofUserLinks(links []*lmodels.UserLink) []*Link {
	resp := make([]*Link, 0, len(links))
	for _, link := range links {
		resp = append(resp, ofUserLink(link))
	}
	return resp
}

type LoginRequest struct {
	Email string `json:"email"`
	// password handled by sha256
	Password string `json:"password"`
}
