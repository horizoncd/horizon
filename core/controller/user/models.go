package user

import (
	"g.hz.netease.com/horizon/pkg/dao/user"
)

type SearchUserResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

func ofUsers(users []user.User) []*SearchUserResponse {
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
