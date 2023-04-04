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
	"strconv"
	"strings"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/dao"
	envregionmodels "github.com/horizoncd/horizon/pkg/models"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

const (
	rootGroupID = 0

	// _parentID one of the field of the group table
	_parentID = "parent_id"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/group/manager/manager_mock.go -package=mock_manager
type GroupManager interface {
	// Create a group
	Create(ctx context.Context, group *envregionmodels.Group) (*envregionmodels.Group, error)
	// Delete a group by id
	Delete(ctx context.Context, id uint) (int64, error)
	// GetByID get a group by id
	GetByID(ctx context.Context, id uint) (*envregionmodels.Group, error)
	// GetByIDs get groups by ids
	GetByIDs(ctx context.Context, ids []uint) ([]*envregionmodels.Group, error)
	// GetByPaths get groups by paths
	GetByPaths(ctx context.Context, paths []string) ([]*envregionmodels.Group, error)
	// GetByNameFuzzily get groups that fuzzily matching the given name
	GetByNameFuzzily(ctx context.Context, name string) ([]*envregionmodels.Group, error)
	// GetByNameFuzzilyIncludeSoftDelete get groups that fuzzily matching the given name
	GetByNameFuzzilyIncludeSoftDelete(ctx context.Context, name string) ([]*envregionmodels.Group, error)
	// GetByIDNameFuzzily get groups that fuzzily matching the given name and id
	GetByIDNameFuzzily(ctx context.Context, id uint, name string) ([]*envregionmodels.Group, error)
	// GetAll return all the groups
	GetAll(ctx context.Context) ([]*envregionmodels.Group, error)
	// UpdateBasic update basic info of a group
	UpdateBasic(ctx context.Context, group *envregionmodels.Group) error
	// GetSubGroupsUnderParentIDs get subgroups under the given parent groups without paging
	GetSubGroupsUnderParentIDs(ctx context.Context, parentIDs []uint) ([]*envregionmodels.Group, error)
	// Transfer move a group under another parent group
	Transfer(ctx context.Context, id, newParentID uint) error
	// GetSubGroups get subgroups of a parent group, order by updateTime desc by default with paging
	GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*envregionmodels.Group, int64, error)
	// GetChildren get children of a parent group, order by updateTime desc by default with paging
	GetChildren(ctx context.Context, parentID uint, pageNumber, pageSize int) ([]*envregionmodels.GroupOrApplication, int64, error)
	// GetByNameOrPathUnderParent get by name or path under a specified parent
	GetByNameOrPathUnderParent(ctx context.Context, name, path string, parentID uint) ([]*envregionmodels.Group, error)
	// GetSubGroupsByGroupIDs get groups and its subGroups by specified groupIDs
	GetSubGroupsByGroupIDs(ctx context.Context, groupIDs []uint) ([]*envregionmodels.Group, error)
	// UpdateRegionSelector update regionSelector
	UpdateRegionSelector(ctx context.Context, id uint, regionSelector string) error
	// GetSelectableRegionsByEnv return selectable regions of the group by environment
	GetSelectableRegionsByEnv(ctx context.Context, id uint, env string) (envregionmodels.RegionParts, error)
	// GetSelectableRegions return selectable regions of the group
	GetSelectableRegions(ctx context.Context, id uint) (envregionmodels.RegionParts, error)
	// GetDefaultRegions return default region the group
	GetDefaultRegions(ctx context.Context, id uint) ([]*envregionmodels.EnvironmentRegion, error)
	// IsRootGroup returns whether it is the root group(groupID equals 0)
	IsRootGroup(ctx context.Context, groupID uint) bool
	// GroupExist returns whether the group exists in db
	GroupExist(ctx context.Context, groupID uint) bool
}

type groupManager struct {
	groupDAO       dao.GroupDAO
	applicationDAO dao.ApplicationDAO
	envregionDAO   dao.EnvironmentRegionDAO
	regionDAO      dao.RegionDAO
}

func NewGroupManager(db *gorm.DB) GroupManager {
	return &groupManager{
		groupDAO:       dao.NewGroupDAO(db),
		applicationDAO: dao.NewApplicationDAO(db),
		envregionDAO:   dao.NewEnvironmentRegionDAO(db),
		regionDAO:      dao.NewRegionDAO(db),
	}
}

func (m groupManager) GetByIDNameFuzzily(ctx context.Context, id uint, name string) ([]*envregionmodels.Group, error) {
	return m.groupDAO.GetByIDNameFuzzily(ctx, id, name)
}

func (m groupManager) GetChildren(ctx context.Context, parentID uint, pageNumber, pageSize int) (
	[]*envregionmodels.GroupOrApplication, int64, error) {
	return m.groupDAO.ListChildren(ctx, parentID, pageNumber, pageSize)
}

func (m groupManager) GetSubGroups(ctx context.Context, id uint,
	pageNumber, pageSize int) ([]*envregionmodels.Group, int64, error) {
	query := formatListGroupQuery(id, pageNumber, pageSize)
	return m.groupDAO.List(ctx, query)
}

func (m groupManager) Transfer(ctx context.Context, id, newParentID uint) error {
	return m.groupDAO.Transfer(ctx, id, newParentID)
}

func (m groupManager) GetByPaths(ctx context.Context, paths []string) ([]*envregionmodels.Group, error) {
	return m.groupDAO.GetByPaths(ctx, paths)
}

func (m groupManager) GetByIDs(ctx context.Context, ids []uint) ([]*envregionmodels.Group, error) {
	return m.groupDAO.GetByIDs(ctx, ids)
}

func (m groupManager) GetByNameFuzzily(ctx context.Context, name string) ([]*envregionmodels.Group, error) {
	return m.groupDAO.GetByNameFuzzily(ctx, name, false)
}

func (m groupManager) GetByNameFuzzilyIncludeSoftDelete(ctx context.Context,
	name string) ([]*envregionmodels.Group, error) {
	return m.groupDAO.GetByNameFuzzily(ctx, name, true)
}

func (m groupManager) GetAll(ctx context.Context) ([]*envregionmodels.Group, error) {
	return m.groupDAO.GetAll(ctx)
}

func (m groupManager) Create(ctx context.Context, group *envregionmodels.Group) (*envregionmodels.Group, error) {
	if err := m.checkApplicationExists(ctx, group); err != nil {
		return nil, err
	}

	id, err := m.groupDAO.Create(ctx, group)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (m groupManager) Delete(ctx context.Context, id uint) (int64, error) {
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

func (m groupManager) GetByID(ctx context.Context, id uint) (*envregionmodels.Group, error) {
	return m.groupDAO.GetByID(ctx, id)
}

func (m groupManager) UpdateBasic(ctx context.Context, group *envregionmodels.Group) error {
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

func (m groupManager) GetSubGroupsUnderParentIDs(ctx context.Context,
	parentIDs []uint) ([]*envregionmodels.Group, error) {
	query := q.New(q.KeyWords{
		_parentID: parentIDs,
	})
	return m.groupDAO.ListWithoutPage(ctx, query)
}

// checkApplicationExists check application is already exists under the same parent
func (m groupManager) checkApplicationExists(ctx context.Context, group *envregionmodels.Group) error {
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

func (m groupManager) GetByNameOrPathUnderParent(ctx context.Context,
	name, path string, parentID uint) ([]*envregionmodels.Group, error) {
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
	splitIds := strings.Split(traversalIDs, ",")
	var ids = make([]uint, len(splitIds))
	for i, id := range splitIds {
		ii, _ := strconv.Atoi(id)
		ids[i] = uint(ii)
	}
	return ids
}

// GetSubGroupsByGroupIDs get groups and its subGroups by specified groupIDs
func (m groupManager) GetSubGroupsByGroupIDs(ctx context.Context, groupIDs []uint) ([]*envregionmodels.Group, error) {
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
	parentsMap := make(map[uint]*envregionmodels.Group, len(parents))
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

func (m groupManager) UpdateRegionSelector(ctx context.Context, id uint, regionSelector string) error {
	return m.groupDAO.UpdateRegionSelector(ctx, id, regionSelector)
}

func (m groupManager) GetSelectableRegionsByEnv(ctx context.Context,
	id uint, env string) (envregionmodels.RegionParts, error) {
	// query regions under env that are not disabled
	envRegionParts, err := m.envregionDAO.ListEnabledRegionsByEnvironment(ctx, env)
	if err != nil {
		return nil, err
	}
	if len(envRegionParts) == 0 {
		return envregionmodels.RegionParts{}, nil
	}

	// get regions with group's regionSelector
	groupRegionParts, err := m.GetSelectableRegions(ctx, id)
	if err != nil {
		return nil, err
	}
	partMap := make(map[string]*envregionmodels.RegionPart)
	for _, p := range groupRegionParts {
		partMap[p.Name] = p
	}

	var regionParts envregionmodels.RegionParts
	for _, p := range envRegionParts {
		if _, ok := partMap[p.Name]; ok {
			regionParts = append(regionParts, p)
		}
	}

	return regionParts, nil
}

func (m groupManager) GetSelectableRegions(ctx context.Context, id uint) (envregionmodels.RegionParts, error) {
	// get regionSelector field from group
	group, err := m.groupDAO.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// unmarshal from yaml to struct
	var regionSelectors envregionmodels.RegionSelectors
	err = yaml.Unmarshal([]byte(group.RegionSelector), &regionSelectors)
	if err != nil {
		return nil, herrors.NewErrGetFailed(herrors.RegionInDB, err.Error())
	}

	regionParts, err := m.regionDAO.ListByRegionSelectors(ctx, regionSelectors)
	if err != nil {
		return nil, err
	}
	if len(regionParts) == 0 {
		return envregionmodels.RegionParts{}, nil
	}

	return regionParts, nil
}

func (m groupManager) GetDefaultRegions(ctx context.Context, id uint) ([]*envregionmodels.EnvironmentRegion, error) {
	selectableRegions, err := m.GetSelectableRegions(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(selectableRegions) == 0 {
		return make([]*envregionmodels.EnvironmentRegion, 0), nil
	}

	selectableRegionMap := make(map[string]*envregionmodels.RegionPart)
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
func (m groupManager) IsRootGroup(ctx context.Context, groupID uint) bool {
	return groupID == rootGroupID
}

// GroupExist returns whether the group exists in db
func (m groupManager) GroupExist(ctx context.Context, groupID uint) bool {
	if m.IsRootGroup(ctx, groupID) {
		return true
	}

	if _, err := m.GetByID(ctx, groupID); err != nil {
		return false
	}
	return true
}
