package user

import (
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/user/models"
)

type SearchUserResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

func ofUser(user *models.User) userauth.User {
	return &userauth.DefaultInfo{
		ID:       user.ID,
		Name:     user.Name,
		FullName: user.FullName,
		Email:    user.Email,
		Admin:    user.Admin,
	}
}

func ofUsers(users []models.User) []*SearchUserResponse {
	resp := make([]*SearchUserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, &SearchUserResponse{
			ID:       u.ID,
			Name:     u.Name,
			FullName: u.FullName,
			Email:    u.Email,
		})
	}
	return resp
}
