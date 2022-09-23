package user

import (
	"time"

	"g.hz.netease.com/horizon/pkg/user/models"
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
