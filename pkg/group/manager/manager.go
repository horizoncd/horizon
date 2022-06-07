package manager

import (
	"context"
	"strconv"
	"strings"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	envregiondao "g.hz.netease.com/horizon/pkg/environmentregion/dao"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	groupdao "g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/group/models"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	regiondao "g.hz.netease.com/horizon/pkg/region/dao"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"github.com/go-yaml/yaml"
)

var (
	// Mgr is the global group manager
	Mgr = New()
)

const (
	// _updateAt one of the field of the group table
	_updateAt = "updated_at"

	// _parentID one of the field of the group table
	_parentID = "parent_id"
)

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

func New() Manager {
	return &manager{
		groupDAO:       groupdao.NewDAO(),
		applicationDAO: applicationdao.NewDAO(),
		envregionDAO:   envregiondao.NewDAO(),
		regionDAO:      regiondao.NewDAO(),
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
	return m.groupDAO.GetByNameFuzzily(ctx, name)
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
	splitIds := strings.Split(traversalIDs, ",")
	var ids = make([]uint, len(splitIds))
	for i, id := range splitIds {
		ii, _ := strconv.Atoi(id)
		ids[i] = uint(ii)
	}
	return ids
}

// GetSubGroupsByGroupIDs get groups and its subGroups by specified groupIDs
func (m manager) GetSubGroupsByGroupIDs(ctx context.Context, groupIDs []uint) ([]*models.Group, error) {
	groupIDSet := make(map[uint]bool)
	for _, groupID := range groupIDs {
		groupIDSet[groupID] = true
	}

	groups, err := m.groupDAO.ListByTraversalIDsContains(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	retGroupIDs := make([]uint, 0)
	for _, group := range groups {
		traversalIDs := FormatIDsFromTraversalIDs(group.TraversalIDs)
		for index, traversalID := range traversalIDs {
			if _, ok := groupIDSet[traversalID]; ok {
				retGroupIDs = append(retGroupIDs, traversalIDs[index:]...)
				break
			}
		}
	}
	return m.groupDAO.GetByIDs(ctx, retGroupIDs)
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

	groupRegionParts, err := m.GetSelectableRegions(ctx, id)
	if err != nil {
		return nil, err
	}

	// get regions under env and regionSelector at the same time
	var regionParts regionmodels.RegionParts
	partMap := make(map[string]*regionmodels.RegionPart)
	for _, p := range envRegionParts {
		partMap[p.Name] = p
	}
	for _, p := range groupRegionParts {
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
	var regionSelectors groupmodels.KubernetesSelectors
	err = yaml.Unmarshal([]byte(group.KubernetesSelector), &regionSelectors)
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
