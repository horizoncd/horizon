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

package environmentregion

import (
	"context"

	environmentregionmanager "github.com/horizoncd/horizon/pkg/manager"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param"
)

type Controller interface {
	ListAll(ctx context.Context) (EnvironmentRegions, error)
	ListByEnvironment(ctx context.Context, environment string) (EnvironmentRegions, error)
	CreateEnvironmentRegion(ctx context.Context, request *CreateEnvironmentRegionRequest) (uint, error)
	SetEnvironmentRegionToDefault(ctx context.Context, id uint) error
	DeleteByID(ctx context.Context, id uint) error
	SetEnvironmentRegionIfAutoFree(ctx context.Context, id uint, autoFree bool) error
}

var _ Controller = (*controller)(nil)

func NewController(param *param.Param) Controller {
	return &controller{
		envRegionMgr: param.EnvRegionMgr,
		regionMgr:    param.RegionMgr,
	}
}

type controller struct {
	envRegionMgr environmentregionmanager.EnvironmentRegionManager
	regionMgr    environmentregionmanager.RegionManager
}

func (c *controller) ListByEnvironment(ctx context.Context, environment string) (EnvironmentRegions, error) {
	environmentRegions, err := c.envRegionMgr.ListByEnvironment(ctx, environment)
	if err != nil {
		return nil, err
	}
	regions, err := c.regionMgr.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return ofRegionModels(regions, environmentRegions), nil
}

// DeleteByID implements Controller
func (c *controller) DeleteByID(ctx context.Context, id uint) error {
	return c.envRegionMgr.DeleteByID(ctx, id)
}

func (c *controller) CreateEnvironmentRegion(ctx context.Context,
	request *CreateEnvironmentRegionRequest) (uint, error) {
	environmentRegion, err := c.envRegionMgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		EnvironmentName: request.EnvironmentName,
		RegionName:      request.RegionName,
		AutoFree:        request.AutoFree,
	})
	if err != nil {
		return 0, err
	}
	return environmentRegion.ID, nil
}

func (c *controller) ListAll(ctx context.Context) (_ EnvironmentRegions, err error) {
	environmentRegions, err := c.envRegionMgr.ListAllEnvironmentRegions(ctx)
	if err != nil {
		return nil, err
	}
	regions, err := c.regionMgr.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	return ofRegionModels(regions, environmentRegions), nil
}

func (c *controller) SetEnvironmentRegionToDefault(ctx context.Context, id uint) error {
	return c.envRegionMgr.SetEnvironmentRegionToDefaultByID(ctx, id)
}

func (c *controller) SetEnvironmentRegionIfAutoFree(ctx context.Context, id uint, autoFree bool) error {
	return c.envRegionMgr.SetEnvironmentRegionIfAutoFree(ctx, id, autoFree)
}
