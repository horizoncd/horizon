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
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	applicationdao "github.com/horizoncd/horizon/pkg/application/dao"
	envregiondao "github.com/horizoncd/horizon/pkg/environmentregion/dao"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	groupdao "github.com/horizoncd/horizon/pkg/group/dao"
	"github.com/horizoncd/horizon/pkg/group/models"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	regiondao "github.com/horizoncd/horizon/pkg/region/dao"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
)

const (
	// _updateAt one of the field of the group table
	_updateAt = "updated_at"

	// _parentID one of the field of the group table
	_parentID = "parent_id"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/group/manager/manager_mock.go -package=mock_manager
type Manager interface {
	// Create a group
	Create(ctx context.Context, group *models.Group) (*models.Group, error)
	// Delete a group by id
	Delete(ctx context.Context, id uint) (int64, error)
	// GetByID get a group by id
	GetByID(ctx context.Context, id uint) (*models.Group, error)
	// GetByIDs get groups by ids
	GetByIDs(ctx context.Context, ids []uint) ([]*models.Group, error)
	// GetByPaths get groups by paths
	GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error)
	// GetByNameFuzzily get groups that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error)
	// GetByNameFuzzilyIncludeSoftDelete get groups that fuzzily matching the given name
	GetByNameFuzzilyIncludeSoftDelete(ctx context.Context, name string) ([]*models.Group, error)
	// GetByIDNameFuzzily get groups that fuzzily matching the given name and id
	GetByIDNameFuzzily(ctx context.Context, id uint, name string) ([]*models.Group, error)
	// GetAll return all the groups
	GetAll(ctx context.Context) ([]*models.Group, error)
	// UpdateBasic update basic info of a group
	UpdateBasic(ctx context.Context, group *models.Group) error
	// GetSubGroupsUnderParentIDs get subgroups under the given parent groups without paging
	GetSubGroupsUnderParentIDs(ctx context.Context, parentIDs []uint) ([]*models.Group, error)
	// Transfer move a group under another parent group
	Transfer(ctx context.Context, id, newParentID uint) error
	// GetSubGroups get subgroups of a parent group, order by updateTime desc by default with paging
	GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*models.Group, int64, error)
	// GetChildren get children of a parent group, order by updateTime desc by default with paging
	GetChildren(ctx context.Context, parentID uint, pageNumber, pageSize int) ([]*models.GroupOrApplication, int64, error)
	// GetByNameOrPathUnderParent get by name or path under a specified parent
	GetByNameOrPathUnderParent(ctx context.Context, name, path string, parentID uint) ([]*models.Group, error)
	// GetSubGroupsByGroupIDs get groups and its subGroups by specified groupIDs
	GetSubGroupsByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Group, error)
	// UpdateRegionSelector update regionSelector
	UpdateRegionSelector(ctx context.Context, id uint, regionSelector string) error
	// GetSelectableRegionsByEnv return selectable regions of the group by environment
	GetSelectableRegionsByEnv(ctx context.Context, id uint, env string) (regionmodels.RegionParts, error)
	// GetSelectableRegions return selectable regions of the group
	GetSelectableRegions(ctx context.Context, id uint) (regionmodels.RegionParts, error)
	// GetDefaultRegions return default region the group
	GetDefaultRegions(ctx context.Context, id uint) ([]*envregionmodels.EnvironmentRegion, error)
	// IsRootGroup returns whether it is the root group(groupID equals 0)
	IsRootGroup(groupID uint) bool
	// GroupExist returns whether the group exists in db
	GroupExist(ctx context.Context, groupID uint) bool
}

type manager struct {
	groupDAO       groupdao.DAO
	applicationDAO applicationdao.DAO
	envregionDAO   envregiondao.DAO
	regionDAO      regiondao.DAO
}

func (m manager) GetByIDNameFuzzily(ctx context.Context, id uint, name string) ([]*models.Group, error) {
	return m.groupDAO.GetByIDNameFuzzily(ctx, id, name)
}

func New(db *gorm.DB) Manager {
	return &manager{
		groupDAO:       groupdao.NewDAO(db),
		applicationDAO: applicationdao.NewDAO(db),
		envregionDAO:   envregiondao.NewDAO(db),
		regionDAO:      regiondao.NewDAO(db),
	}
}

func (m manager) GetChildren(ctx context.Context, parentID uint, pageNumber, pageSize int) (
	[]*models.GroupOrApplication, int64, error) {
	return m.groupDAO.ListChildren(ctx, parentID, pageNumber, pageSize)
}

func (m manager) GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*models.Group, int64, error) {
	query := formatListGroupQuery(id, pageNumber, pageSize)
	return m.groupDAO.List(ctx, query)
}

func (m manager) Transfer(ctx context.Context, id, newParentID uint) error {
	return m.groupDAO.Transfer(ctx, id, newParentID)
}

func (m manager) GetByPaths(ctx context.Context, paths []string) ([]*models.Group, error) {
	return m.groupDAO.GetByPaths(ctx, paths)
}

func (m manager) GetByIDs(ctx context.Context, ids []uint) ([]*models.Group, error) {
	return m.groupDAO.GetByIDs(ctx, ids)
}

func (m manager) GetByNameFuzzily(ctx context.Context, name string) ([]*models.Group, error) {
	return m.groupDAO.GetByNameFuzzily(ctx, name, false)
}

func (m manager) GetByNameFuzzilyIncludeSoftDelete(ctx context.Context, name string) ([]*models.Group, error) {
	return m.groupDAO.GetByNameFuzzily(ctx, name, true)
}

func (m manager) GetAll(ctx context.Context) ([]*models.Group, error) {
	return m.groupDAO.GetAll(ctx)
}

func (m manager) Create(ctx context.Context, group *models.Group) (*models.Group, error) {
	if err := m.checkApplicationExists(ctx, group); err != nil {
		return nil, err
	}

	id, err := m.groupDAO.Create(ctx, group)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (m manager) Delete(ctx context.Context, id uint) (int64, error) {
	count, err := m.groupDAO.CountByParentID(ctx, id)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, herrors.ErrGroupHasChildren
	}

	count, err = m.applicationDAO.CountByGroupID(ctx, id)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, herrors.ErrGroupHasChildren
	}

	return m.groupDAO.Delete(ctx, id)
}

func (m manager) GetByID(ctx context.Context, id uint) (*models.Group, error) {
	return m.groupDAO.GetByID(ctx, id)
}

func (m manager) UpdateBasic(ctx context.Context, group *models.Group) error {
	if err := m.checkApplicationExists(ctx, group); err != nil {
		return err
	}

	// check record exist
	_, err := m.groupDAO.GetByID(ctx, group.ID)
	if err != nil {
		return err
	}

	// check if there's record with the same parentID and name
	err = m.groupDAO.CheckNameUnique(ctx, group)
	if err != nil {
		return err
	}
	// check if there's a record with the same parentID and path
	err = m.groupDAO.CheckPathUnique(ctx, group)
	if err != nil {
		return err
	}

	return m.groupDAO.UpdateBasic(ctx, group)
}

func (m manager) GetSubGroupsUnderParentIDs(ctx context.Context, parentIDs []uint) ([]*models.Group, error) {
	query := q.New(q.KeyWords{
		_parentID: parentIDs,
	})
	return m.groupDAO.ListWithoutPage(ctx, query)
}

// checkApplicationExists check application is already exists under the same parent
func (m manager) checkApplicationExists(ctx context.Context, group *models.Group) error {
	apps, err := m.applicationDAO.GetByNamesUnderGroup(ctx,
		group.ParentID, []string{group.Name, group.Path})
	if err != nil {
		return err
	}
	if len(apps) > 0 {
		return herrors.ErrGroupConflictWithApplication
	}
	return nil
}

func (m manager) GetByNameOrPathUnderParent(ctx context.Context,
	name, path string, parentID uint) ([]*models.Group, error) {
	return m.groupDAO.GetByNameOrPathUnderParent(ctx, name, path, parentID)
}

// formatListGroupQuery query info for listing groups under a parent group, order by updated_at desc by default
func formatListGroupQuery(id uint, pageNumber, pageSize int) *q.Query {
	query := q.New(q.KeyWords{
		_parentID: id,
	})
	query.PageNumber = pageNumber
	query.PageSize = pageSize
	// sort by updated_at desc default，let newer items be in head
	s := q.NewSort(_updateAt, true)
	query.Sorts = []*q.Sort{s}

	return query
}

// FormatIDsFromTraversalIDs format id array from traversalIDs(1,2,3)
func FormatIDsFromTraversalIDs(traversalIDs string) []uint {
	ids, _ := common.UnmarshalTraversalIDS(traversalIDs)
	return ids
}

// GetSubGroupsByGroupIDs get groups and its subGroups by specified groupIDs
func (m manager) GetSubGroupsByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Group, error) {
	IDs := make([]uint, 0)
	groupIDSet := make(map[uint]struct{})
	for _, groupID := range groupIDs {
		if _, ok := groupIDSet[groupID]; !ok {
			groupIDSet[groupID] = struct{}{}
			IDs = append(IDs, groupID)
		}
	}

	parents, err := m.groupDAO.GetByIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	parentsMap := make(map[uint]*models.Group, len(parents))
	for _, parent := range parents {
		parentsMap[parent.ID] = parent
	}

	children, err := m.groupDAO.ListByTraversalIDsContains(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		if _, ok := groupIDSet[child.ID]; !ok {
			IDs = append(IDs, child.ID)
			groupIDSet[child.ID] = struct{}{}
		}
		parent := parentsMap[child.ParentID]
		if parent == nil {
			continue
		}
		traversalIDs := FormatIDsFromTraversalIDs(
			strings.TrimPrefix(
				child.TraversalIDs, fmt.Sprintf("%s,", parent.TraversalIDs),
			),
		)
		for _, traversalID := range traversalIDs {
			if _, ok := groupIDSet[traversalID]; !ok {
				IDs = append(IDs, traversalID)
				groupIDSet[traversalID] = struct{}{}
			}
		}
	}
	return m.groupDAO.GetByIDs(ctx, IDs)
}

func (m manager) UpdateRegionSelector(ctx context.Context, id uint, regionSelector string) error {
	return m.groupDAO.UpdateRegionSelector(ctx, id, regionSelector)
}

func (m manager) GetSelectableRegionsByEnv(ctx context.Context, id uint, env string) (regionmodels.RegionParts, error) {
	// query regions under env that are not disabled
	envRegionParts, err := m.envregionDAO.ListEnabledRegionsByEnvironment(ctx, env)
	if err != nil {
		return nil, err
	}
	if len(envRegionParts) == 0 {
		return regionmodels.RegionParts{}, nil
	}

	// get regions with group's regionSelector
	groupRegionParts, err := m.GetSelectableRegions(ctx, id)
	if err != nil {
		return nil, err
	}
	partMap := make(map[string]*regionmodels.RegionPart)
	for _, p := range groupRegionParts {
		partMap[p.Name] = p
	}

	var regionParts regionmodels.RegionParts
	for _, p := range envRegionParts {
		if _, ok := partMap[p.Name]; ok {
			regionParts = append(regionParts, p)
		}
	}

	return regionParts, nil
}

func (m manager) GetSelectableRegions(ctx context.Context, id uint) (regionmodels.RegionParts, error) {
	// get regionSelector field from group
	group, err := m.groupDAO.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// unmarshal from yaml to struct
	var regionSelectors groupmodels.RegionSelectors
	err = yaml.Unmarshal([]byte(group.RegionSelector), &regionSelectors)
	if err != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, err.Error())
	}

	regionParts, err := m.regionDAO.ListByRegionSelectors(ctx, regionSelectors)
	if err != nil {
		return nil, err
	}
	if len(regionParts) == 0 {
		return regionmodels.RegionParts{}, nil
	}

	return regionParts, nil
}

func (m manager) GetDefaultRegions(ctx context.Context, id uint) ([]*envregionmodels.EnvironmentRegion, error) {
	selectableRegions, err := m.GetSelectableRegions(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(selectableRegions) == 0 {
		return make([]*envregionmodels.EnvironmentRegion, 0), nil
	}

	selectableRegionMap := make(map[string]*regionmodels.RegionPart)
	for _, region := range selectableRegions {
		selectableRegionMap[region.Name] = region
	}

	envDefaultRegions, err := m.envregionDAO.GetDefaultRegions(ctx)
	if err != nil {
		return nil, err
	}
	if len(envDefaultRegions) == 0 {
		return make([]*envregionmodels.EnvironmentRegion, 0), nil
	}

	var res []*envregionmodels.EnvironmentRegion
	for _, region := range envDefaultRegions {
		if _, ok := selectableRegionMap[region.RegionName]; ok {
			res = append(res, region)
		}
	}

	return res, nil
}

// IsRootGroup return whether it is the root group(groupID equals 0)
func (m manager) IsRootGroup(groupID uint) bool {
	return groupID == models.RootGroupID
}

// GroupExist returns whether the group exists in db
func (m manager) GroupExist(ctx context.Context, groupID uint) bool {
	if m.IsRootGroup(groupID) {
		return true
	}

	if _, err := m.GetByID(ctx, groupID); err != nil {
		return false
	}
	return true
}
