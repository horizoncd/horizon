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

package environment

import (
	"time"

	"github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/service"
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
