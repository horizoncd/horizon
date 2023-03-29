package environment

import (
	"time"

	"github.com/horizoncd/horizon/pkg/environment/models"
	"github.com/horizoncd/horizon/pkg/environment/service"
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
func ofEnvironmentModels(envs []*models.Environment, autoFreeSvc *service.AutoFreeSVC) Environments {
	environments := make(Environments, 0)
	for _, env := range envs {
		environments = append(environments, ofEnvironmentModel(env, autoFreeSvc.WhetherSupported(env.Name)))
	}
	return environments
}

func ofEnvironmentModel(env *models.Environment, isAutoFree bool) *Environment {
	return &Environment{
		ID:          env.ID,
		Name:        env.Name,
		DisplayName: env.DisplayName,
		AutoFree:    isAutoFree,
		CreatedAt:   env.CreatedAt,
		UpdatedAt:   env.UpdatedAt,
	}
}

type CreateEnvironmentRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type UpdateEnvironmentRequest struct {
	DisplayName string `json:"displayName"`
}
