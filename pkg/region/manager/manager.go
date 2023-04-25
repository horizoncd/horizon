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

package manager

import (
	"context"

	"github.com/horizoncd/horizon/core/common"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	regiondao "github.com/horizoncd/horizon/pkg/region/dao"
	"github.com/horizoncd/horizon/pkg/region/models"
	registrydao "github.com/horizoncd/horizon/pkg/registry/dao"
	registrymodels "github.com/horizoncd/horizon/pkg/registry/models"
	tagdao "github.com/horizoncd/horizon/pkg/tag/dao"
	"gorm.io/gorm"
)

type Manager interface {
	// Create a region
	Create(ctx context.Context, region *models.Region) (*models.Region, error)
	// ListAll list all regions
	ListAll(ctx context.Context) ([]*models.Region, error)
	// ListRegionEntities list all region entity
	ListRegionEntities(ctx context.Context) ([]*models.RegionEntity, error)
	// GetRegionEntity get region entity
	GetRegionEntity(ctx context.Context, regionName string) (*models.RegionEntity, error)
	GetRegionByID(ctx context.Context, id uint) (*models.RegionEntity, error)
	GetRegionByName(ctx context.Context, name string) (*models.Region, error)
	// UpdateByID update region by id
	UpdateByID(ctx context.Context, id uint, region *models.Region) error
	// ListByRegionSelectors list region by tags
	ListByRegionSelectors(ctx context.Context, selectors groupmodels.RegionSelectors) (models.RegionParts, error)
	// DeleteByID delete region by id
	DeleteByID(ctx context.Context, id uint) error
}

type manager struct {
	regionDAO   regiondao.DAO
	registryDAO registrydao.DAO
	tagDAO      tagdao.DAO
}

func New(db *gorm.DB) Manager {
	return &manager{
		regionDAO:   regiondao.NewDAO(db),
		registryDAO: registrydao.NewDAO(db),
		tagDAO:      tagdao.NewDAO(db),
	}
}

func (m *manager) Create(ctx context.Context, region *models.Region) (*models.Region, error) {
	return m.regionDAO.Create(ctx, region)
}

func (m *manager) ListAll(ctx context.Context) ([]*models.Region, error) {
	return m.regionDAO.ListAll(ctx)
}

func (m *manager) ListRegionEntities(ctx context.Context) (ret []*models.RegionEntity, err error) {
	var regions []*models.Region
	regions, err = m.regionDAO.ListAll(ctx)
	if err != nil {
		return
	}

	for _, region := range regions {
		tags, err := m.tagDAO.ListByResourceTypeID(ctx, common.ResourceRegion, region.ID)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &models.RegionEntity{
			Region: region,
			Tags:   tags,
		})
	}
	return
}

func (m *manager) GetRegionEntity(ctx context.Context,
	regionName string) (*models.RegionEntity, error) {
	region, err := m.regionDAO.GetRegion(ctx, regionName)
	if err != nil {
		return nil, err
	}

	registry, err := m.getRegistryByRegion(ctx, region)
	if err != nil {
		return nil, err
	}

	return &models.RegionEntity{
		Region:   region,
		Registry: registry,
	}, nil
}

func (m *manager) UpdateByID(ctx context.Context, id uint, region *models.Region) error {
	_, err := m.getRegistryByRegion(ctx, region)
	if err != nil {
		return err
	}
	// todo do more filed validation, for example ingressDomain must be format of the domain name
	return m.regionDAO.UpdateByID(ctx, id, region)
}

func (m *manager) getRegistryByRegion(ctx context.Context, region *models.Region) (*registrymodels.Registry, error) {
	registry, err := m.registryDAO.GetByID(ctx, region.RegistryID)
	if err != nil {
		return nil, err
	}
	return registry, nil
}

func (m *manager) ListByRegionSelectors(ctx context.Context, selectors groupmodels.RegionSelectors) (
	models.RegionParts, error) {
	return m.regionDAO.ListByRegionSelectors(ctx, selectors)
}

func (m *manager) DeleteByID(ctx context.Context, id uint) error {
	return m.regionDAO.DeleteByID(ctx, id)
}

func (m *manager) GetRegionByID(ctx context.Context, id uint) (*models.RegionEntity, error) {
	region, err := m.regionDAO.GetRegionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	registry, err := m.getRegistryByRegion(ctx, region)
	if err != nil {
		return nil, err
	}

	tags, err := m.tagDAO.ListByResourceTypeID(ctx, common.ResourceRegion, region.ID)
	if err != nil {
		return nil, err
	}

	return &models.RegionEntity{
		Region:   region,
		Registry: registry,
		Tags:     tags,
	}, nil
}

func (m *manager) GetRegionByName(ctx context.Context, name string) (*models.Region, error) {
	return m.regionDAO.GetRegion(ctx, name)
}
