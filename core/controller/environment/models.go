package environment

import (
	"time"

	"g.hz.netease.com/horizon/pkg/environment/models"
)

type Environment struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	AutoFree    bool      `json:"autoFree"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Environments []*Environment

// ofEnvironmentModels []*models.Environment to []*Environment
func ofEnvironmentModels(envs []*models.Environment) Environments {
	environments := make(Environments, 0)
	for _, env := range envs {
		environments = append(environments, ofEnvironmentModel(env))
	}
	return environments
}

func ofEnvironmentModel(env *models.Environment) *Environment {
	return &Environment{
		ID:          env.ID,
		Name:        env.Name,
		DisplayName: env.DisplayName,
		AutoFree:    env.AutoFree,
		CreatedAt:   env.CreatedAt,
		UpdatedAt:   env.UpdatedAt,
	}
}

type CreateEnvironmentRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	AutoFree    bool   `json:"autoFree"`
}

type UpdateEnvironmentRequest struct {
	DisplayName string `json:"displayName"`
	AutoFree    bool   `json:"autoFree"`
}
