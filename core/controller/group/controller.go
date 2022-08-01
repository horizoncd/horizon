package group

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	"g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/group/service"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	tmanager "g.hz.netease.com/horizon/pkg/template/manager"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"github.com/go-yaml/yaml"
)

const (
	// ErrCodeNotFound a kind of error code, returned when there's no group matching the given id
	ErrCodeNotFound = errors.ErrorCode("RecordNotFound")
	// ErrGroupHasChildren a kind of error code, returned when deleting a group which still has some children
	ErrGroupHasChildren = errors.ErrorCode("GroupHasChildren")
)

type Controller interface {
	// CreateGroup add a group
	CreateGroup(ctx context.Context, newGroup *NewGroup) (uint, error)
	// Delete remove a group by the id
	Delete(ctx context.Context, id uint) error
	// GetByID get a group by the id
	GetByID(ctx context.Context, id uint) (*StructuredGroup, error)
	// GetByFullPath get a group by the URLPath
	GetByFullPath(ctx context.Context, resourcePath string, resourceType string) (*service.Child, error)
	// Transfer put a group under another parent group
	Transfer(ctx context.Context, id, newParentID uint) error
	// UpdateBasic update basic info of a group, including name, path, description and visibilityLevel
	UpdateBasic(ctx context.Context, id uint, updateGroup *UpdateGroup) error
	// GetSubGroups get subgroups of a group
	GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) ([]*service.Child, int64, error)
	// GetChildren get children of a group, including subgroups and applications
	GetChildren(ctx context.Context, id uint, pageNumber, pageSize int) ([]*service.Child, int64, error)
	// SearchGroups search subGroups of a group
	SearchGroups(ctx context.Context, params *SearchParams) ([]*service.Child, int64, error)
	// SearchChildren search children of a group, including subgroups and applications
	SearchChildren(ctx context.Context, params *SearchParams) ([]*service.Child, int64, error)
	// ListAuthedGroup get all the authed groups of current user(if is admin, return all the groups)
	ListAuthedGroup(ctx context.Context) ([]*Group, error)
	// UpdateRegionSelector update regionSelector
	UpdateRegionSelector(ctx context.Context, id uint, regionSelector RegionSelectors) error
}

type controller struct {
	groupManager       groupmanager.Manager
	applicationManager appmanager.Manager
	clusterManager     clustermanager.Manager
	memberSvc          memberservice.Service
	templateMgr        tmanager.Manager
	templateReleaseMgr trmanager.Manager
}

// NewController initializes a new group controller
func NewController(param *param.Param) Controller {
	return &controller{
		groupManager:       param.GroupManager,
		applicationManager: param.ApplicationManager,
		clusterManager:     param.ClusterMgr,
		memberSvc:          param.MemberService,
		templateMgr:        param.TemplateMgr,
		templateReleaseMgr: param.TemplateReleaseManager,
	}
}

// GetChildren get children of a group, including subgroups and applications
func (c *controller) GetChildren(ctx context.Context, id uint, pageNumber, pageSize int) (
	[]*service.Child, int64, error) {
	var parent *models.Group
	var full *service.Full
	if id > 0 {
		var err error
		parent, err = c.groupManager.GetByID(ctx, id)
		if err != nil {
			return nil, 0, err
		}

		full, err = c.formatFullFromGroup(ctx, parent)
		if err != nil {
			return nil, 0, err
		}
	}

	// query children
	children, count, err := c.groupManager.GetChildren(ctx, id, pageNumber, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// calculate childrenCount
	parentIDs := make([]uint, len(children))
	for i, g := range children {
		if g.Type == service.ChildTypeGroup {
			parentIDs[i] = g.ID
		}
	}
	groups, err := c.groupManager.GetSubGroupsUnderParentIDs(ctx, parentIDs)
	if err != nil {
		return nil, 0, err
	}
	childrenCountMap := map[uint]int{}
	for _, g := range groups {
		childrenCountMap[g.ParentID]++
	}

	// format GroupChild
	var gChildren = make([]*service.Child, len(children))
	for i, val := range children {
		var fName, fPath string
		if id == 0 {
			fName = val.Name
			fPath = fmt.Sprintf("/%s", val.Path)
		} else {
			fName = fmt.Sprintf("%s/%s", full.FullName, val.Name)
			fPath = fmt.Sprintf("%s/%s", full.FullPath, val.Path)
		}
		child := service.ConvertGroupOrApplicationToChild(val, &service.Full{
			FullName: fName,
			FullPath: fPath,
		})
		child.ChildrenCount = childrenCountMap[child.ID]

		gChildren[i] = child
	}

	return gChildren, count, nil
}

// SearchGroups search subGroups of a group
func (c *controller) SearchGroups(ctx context.Context, params *SearchParams) ([]*service.Child, int64, error) {
	if params.Filter == "" {
		return c.GetSubGroups(ctx, params.GroupID, params.PageNumber, params.PageSize)
	}

	// query groups by the name fuzzily
	var matchedGroups []*models.Group
	var err error
	if params.GroupID == 0 {
		matchedGroups, err = c.groupManager.GetByNameFuzzily(ctx, params.Filter)
	} else {
		matchedGroups, err = c.groupManager.GetByIDNameFuzzily(ctx, params.GroupID, params.Filter)
	}
	if err != nil {
		return nil, 0, err
	}
	if matchedGroups == nil {
		return []*service.Child{}, 0, nil
	}

	// query groups in ids (split matchedGroups' traversalIDs by ',')
	groups, err := c.formatGroupsInTraversalIDs(ctx, matchedGroups)
	if err != nil {
		return nil, 0, err
	}

	// generate children with level struct
	childrenWithLevelStruct := generateChildrenWithLevelStruct(params.GroupID, groups, []*appmodels.Application{})

	// sort children by updatedAt desc
	sort.SliceStable(childrenWithLevelStruct, func(i, j int) bool {
		return childrenWithLevelStruct[i].UpdatedAt.After(childrenWithLevelStruct[j].UpdatedAt)
	})
	return childrenWithLevelStruct, int64(len(childrenWithLevelStruct)), nil
}

// SearchChildren search children of a group, including subgroups and applications
func (c *controller) SearchChildren(ctx context.Context, params *SearchParams) ([]*service.Child, int64, error) {
	if params.Filter == "" {
		return c.GetChildren(ctx, params.GroupID, params.PageNumber, params.PageSize)
	}

	// query groups by the name fuzzily
	var matchedGroups []*models.Group
	var err error
	if params.GroupID == 0 {
		matchedGroups, err = c.groupManager.GetByNameFuzzily(ctx, params.Filter)
	} else {
		matchedGroups, err = c.groupManager.GetByIDNameFuzzily(ctx, params.GroupID, params.Filter)
	}
	if err != nil {
		return nil, 0, err
	}

	// query applications by the name fuzzily
	matchedApplications, err := c.applicationManager.GetByNameFuzzily(ctx, params.Filter)
	if err != nil {
		return nil, 0, err
	}
	var groupIDs []uint
	for _, application := range matchedApplications {
		groupIDs = append(groupIDs, application.GroupID)
	}
	groups, err := c.groupManager.GetByIDs(ctx, groupIDs)
	if err != nil {
		return nil, 0, err
	}

	matchedGroups = append(matchedGroups, groups...)
	// query groups in ids (split matchedGroups' traversalIDs by ',')
	groups, err = c.formatGroupsInTraversalIDs(ctx, matchedGroups)
	if err != nil {
		return nil, 0, err
	}

	// generate children with level struct
	childrenWithLevelStruct := generateChildrenWithLevelStruct(params.GroupID, groups, matchedApplications)

	// sort children by updatedAt desc
	sort.SliceStable(childrenWithLevelStruct, func(i, j int) bool {
		return childrenWithLevelStruct[i].UpdatedAt.After(childrenWithLevelStruct[j].UpdatedAt)
	})
	return childrenWithLevelStruct, int64(len(childrenWithLevelStruct)), nil
}

// GetSubGroups get subgroups of a group
func (c *controller) GetSubGroups(ctx context.Context, id uint, pageNumber, pageSize int) (
	[]*service.Child, int64, error) {
	var parent *models.Group
	var full *service.Full
	if id > 0 {
		var err error
		parent, err = c.groupManager.GetByID(ctx, id)
		if err != nil {
			return nil, 0, err
		}

		full, err = c.formatFullFromGroup(ctx, parent)
		if err != nil {
			return nil, 0, err
		}
	}

	// query subGroups
	subGroups, count, err := c.groupManager.GetSubGroups(ctx, id, pageNumber, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// calculate childrenCount
	parentIDs := make([]uint, len(subGroups))
	for i, g := range subGroups {
		parentIDs[i] = g.ID
	}
	groups, err := c.groupManager.GetSubGroupsUnderParentIDs(ctx, parentIDs)
	if err != nil {
		return nil, 0, err
	}
	childrenCountMap := map[uint]int{}
	for _, g := range groups {
		childrenCountMap[g.ParentID]++
	}

	// format GroupChild
	var gChildren = make([]*service.Child, len(subGroups))
	for i, s := range subGroups {
		var fName, fPath string
		if id == 0 {
			fName = s.Name
			fPath = fmt.Sprintf("/%s", s.Path)
		} else {
			fName = fmt.Sprintf("%s/%s", full.FullName, s.Name)
			fPath = fmt.Sprintf("%s/%s", full.FullPath, s.Path)
		}
		child := service.ConvertGroupToChild(s, &service.Full{
			FullName: fName,
			FullPath: fPath,
		})
		child.ChildrenCount = childrenCountMap[child.ID]

		gChildren[i] = child
	}

	return gChildren, count, nil
}

// UpdateBasic update basic info of a group, including name, path, description and visibilityLevel
func (c *controller) UpdateBasic(ctx context.Context, id uint, updateGroup *UpdateGroup) error {
	group := convertUpdateGroupToGroup(updateGroup)
	group.ID = id

	err := c.groupManager.UpdateBasic(ctx, group)
	if err != nil {
		return err
	}

	return nil
}

// Transfer put a group under another parent group
func (c *controller) Transfer(ctx context.Context, id, newParentID uint) error {
	err := c.groupManager.Transfer(ctx, id, newParentID)
	if err != nil {
		return err
	}

	return nil
}

// CreateGroup add a group
func (c *controller) CreateGroup(ctx context.Context, newGroup *NewGroup) (uint, error) {
	groupEntity := convertNewGroupToGroup(newGroup)

	group, err := c.groupManager.Create(ctx, groupEntity)
	if err != nil {
		return 0, err
	}

	return group.ID, err
}

// GetByFullPath get a group by the URLPath
func (c *controller) GetByFullPath(ctx context.Context,
	resourcePath string, resourceType string) (*service.Child, error) {
	const op = "get record by fullPath"

	var errNotMatch = errors.E(op, http.StatusNotFound, ErrCodeNotFound,
		fmt.Sprintf("no resource matching the resourcePath: %s", resourcePath))

	if len(resourcePath) == 0 {
		return nil, errNotMatch
	}

	if resourceType == "" {
		// resourcePath: /a/b => {a, b}
		paths := strings.Split(resourcePath[1:], "/")
		groups, err := c.groupManager.GetByPaths(ctx, paths)
		if err != nil {
			return nil, err
		}

		// get mapping between id and group
		idToGroup := service.GenerateIDToGroup(groups)

		// get mapping between id and full
		idToFull := service.GenerateIDToFull(groups)

		// 1. match group
		for k, v := range idToFull {
			// resourcePath pointing to a group
			if v.FullPath == resourcePath {
				g := idToGroup[k]
				child := service.ConvertGroupToChild(g, v)
				return child, nil
			}
		}

		// 2. match application
		if len(paths) < 2 {
			return nil, errNotMatch
		}
		app, err := c.applicationManager.GetByName(ctx, paths[len(paths)-1])
		if app != nil && err == nil {
			appParentFull, ok := idToFull[app.GroupID]
			if ok && fmt.Sprintf("%s/%s", appParentFull.FullPath, app.Name) == resourcePath {
				return service.ConvertApplicationToChild(app, &service.Full{
					FullName: fmt.Sprintf("%s/%s", appParentFull.FullName, app.Name),
					FullPath: fmt.Sprintf("%s/%s", appParentFull.FullPath, app.Name),
				}), nil
			}
		}

		// 3. match cluster
		if len(paths) < 3 {
			return nil, errNotMatch
		}
		cluster, err := c.clusterManager.GetByName(ctx, paths[len(paths)-1])
		if err != nil || cluster == nil {
			return nil, errNotMatch
		}
		app, err = c.applicationManager.GetByID(ctx, cluster.ApplicationID)
		if err != nil {
			return nil, errNotMatch
		}
		appParentFull, ok := idToFull[app.GroupID]
		if ok && fmt.Sprintf("%s/%s/%s", appParentFull.FullPath, app.Name, cluster.Name) == resourcePath {
			return service.ConvertClusterToChild(cluster, &service.Full{
				FullName: fmt.Sprintf("%s/%s/%s", appParentFull.FullName, app.Name, cluster.Name),
				FullPath: fmt.Sprintf("%s/%s/%s", appParentFull.FullPath, app.Name, cluster.Name),
			}), nil
		}

		return nil, errNotMatch
	}

	pathArr := strings.Split(strings.TrimPrefix(resourcePath, "/"), "/")
	// for /group1/group2/group3/template1
	if resourceType == common.ResourceTemplate {
		if len(pathArr) < 1 {
			return nil, errNotMatch
		}
		templateName := pathArr[len(pathArr)-1]
		template, err := c.templateMgr.GetByName(ctx, templateName)
		if err != nil {
			return nil, errNotMatch
		}

		if template.GroupID == service.RootGroupID {
			if len(pathArr) == 1 {
				return service.ConvertTemplateToChild(template, &service.Full{
					FullName: template.Name,
					FullPath: fmt.Sprintf("/%s", template.Name),
				}), nil
			}
			return nil, errNotMatch
		}

		groups, err := c.groupManager.GetByPaths(ctx, pathArr[:len(pathArr)-1])
		if err != nil {
			return nil, err
		}

		// get mapping between id and full
		idToFull := service.GenerateIDToFull(groups)
		full, ok := idToFull[template.GroupID]
		if ok && fmt.Sprintf("%s/%s", full.FullPath, template.Name) == resourcePath {
			return service.ConvertTemplateToChild(template, &service.Full{
				FullName: fmt.Sprintf("%s/%s", full.FullName, template.Name),
				FullPath: fmt.Sprintf("%s/%s", full.FullPath, template.Name),
			}), nil
		}
	} else if resourceType == common.ResourceTemplateRelease {
		// for /template1/release1
		if len(pathArr) < 2 {
			return nil, errNotMatch
		}

		templateName := pathArr[len(pathArr)-2]
		template, err := c.templateMgr.GetByName(ctx, templateName)
		if err != nil {
			return nil, err
		}

		releaseName := pathArr[len(pathArr)-1]
		release, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
		if err != nil {
			return nil, errNotMatch
		}

		if template.GroupID == service.RootGroupID {
			if len(pathArr) == 2 {
				return service.ConvertReleaseToChild(release, &service.Full{
					FullName: fmt.Sprintf("%s/%s", template.Name, releaseName),
					FullPath: fmt.Sprintf("/%s/%s", template.Name, releaseName),
				}), nil
			}
			return nil, errNotMatch
		}

		groups, err := c.groupManager.GetByPaths(ctx, pathArr[:len(pathArr)-2])
		if err != nil {
			return nil, err
		}

		// get mapping between id and full
		idToFull := service.GenerateIDToFull(groups)
		full, ok := idToFull[template.GroupID]
		if ok && fmt.Sprintf("%s/%s/%s", full.FullPath, template.Name, releaseName) == resourcePath {
			return service.ConvertReleaseToChild(release, &service.Full{
				FullName: fmt.Sprintf("%s/%s/%s", full.FullName, template.Name, releaseName),
				FullPath: fmt.Sprintf("%s/%s/%s", full.FullPath, template.Name, releaseName),
			}), nil
		}
	}
	return nil, errNotMatch
}

func (c *controller) ListAuthedGroup(ctx context.Context) ([]*Group, error) {
	currenUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	groups, err := c.groupManager.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	var authedGroups []*models.Group
	if currenUser.IsAdmin() {
		authedGroups = groups
	} else {
		authedGroups = make([]*models.Group, 0)
		for _, item := range groups {
			// TODO: get all group member in one request
			strID := strconv.FormatUint(uint64(item.ID), 10)
			member, err := c.memberSvc.GetMemberOfResource(ctx, common.ResourceGroup, strID)
			if err != nil {
				return nil, err
			}
			if member != nil && (member.Role == role.Owner ||
				member.Role == role.Maintainer || member.Role == role.PE) {
				authedGroups = append(authedGroups, item)
			}
		}
	}
	return c.ofGroupModel(ctx, authedGroups)
}

// GetByID get a group by the id
func (c *controller) GetByID(ctx context.Context, id uint) (*StructuredGroup, error) {
	group, err := c.groupManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	full, err := c.formatFullFromGroup(ctx, group)
	if err != nil {
		return nil, err
	}

	var regionSelectors RegionSelectors
	err = yaml.Unmarshal([]byte(group.RegionSelector), &regionSelectors)
	if err != nil {
		return nil, err
	}

	return &StructuredGroup{
		Group: &Group{
			ID:              group.ID,
			Name:            group.Name,
			Path:            group.Path,
			VisibilityLevel: group.VisibilityLevel,
			Description:     group.Description,
			ParentID:        group.ParentID,
			TraversalIDs:    group.TraversalIDs,
			UpdatedAt:       group.UpdatedAt,
			FullName:        full.FullName,
			FullPath:        full.FullPath,
		},
		RegionSelectors: regionSelectors,
	}, nil
}

// Delete remove a group by the id
func (c *controller) Delete(ctx context.Context, id uint) error {
	const op = "group *controller: delete group by id"

	rowsAffected, err := c.groupManager.Delete(ctx, id)
	if err != nil {
		if err == herrors.ErrGroupHasChildren {
			return errors.E(op, http.StatusBadRequest, ErrGroupHasChildren, herrors.ErrGroupHasChildren)
		}
		return errors.E(op, fmt.Sprintf("failed to delete the group matching the id: %d", id), err)
	}
	if rowsAffected == 0 {
		return errors.E(op, http.StatusNotFound, ErrCodeNotFound, fmt.Sprintf("no group matching the id: %d", id))
	}

	return nil
}

// formatGroupsInTraversalIDs query groups by ids (split traversalIDs by ',')
func (c *controller) formatGroupsInTraversalIDs(ctx context.Context, groups []*models.Group) ([]*models.Group, error) {
	var ids []uint
	for _, g := range groups {
		ids = append(ids, groupmanager.FormatIDsFromTraversalIDs(g.TraversalIDs)...)
	}

	groupsByIDs, err := c.groupManager.GetByIDs(ctx, ids)
	if err != nil {
		return []*models.Group{}, err
	}

	return groupsByIDs, nil
}

// generateChildrenWithLevelStruct generate subgroups with level struct
func generateChildrenWithLevelStruct(groupID uint, groups []*models.Group,
	applications []*appmodels.Application) []*service.Child {
	// get mapping between id and full
	idToFull := service.GenerateIDToFull(groups)

	// record the mapping between parentID and children
	parentIDToChildren := make(map[uint][]*service.Child)

	// first level children under the group
	firstLevelChildren := make([]*service.Child, 0)

	// generate children by applications
	for _, application := range applications {
		parent := idToFull[application.GroupID]
		child := service.ConvertApplicationToChild(application, &service.Full{
			FullName: fmt.Sprintf("%s/%s", parent.FullName, application.Name),
			FullPath: fmt.Sprintf("%s/%s", parent.FullPath, application.Name),
		})
		if application.GroupID == groupID {
			firstLevelChildren = append(firstLevelChildren, child)
		}
		parentIDToChildren[application.GroupID] = append(parentIDToChildren[application.GroupID], child)
	}

	// reverse the order
	sort.Sort(sort.Reverse(models.Groups(groups)))
	for _, g := range groups {
		// get fullName and fullPath by id
		full := idToFull[g.ID]
		child := service.ConvertGroupToChild(g, full)

		// record children of the group whose id is g.parentID
		parentIDToChildren[g.ParentID] = append(parentIDToChildren[g.ParentID], child)

		if v, ok := parentIDToChildren[g.ID]; ok {
			// sort children by type, 'group' in the front of the array while 'application' in the back
			sort.SliceStable(v, func(i, j int) bool {
				return strings.Compare(v[i].Type, v[j].Type) > 0
			})

			child.ChildrenCount = len(v)
			child.Children = v
		}

		if g.ParentID == groupID {
			firstLevelChildren = append(firstLevelChildren, child)
		}
	}

	return firstLevelChildren
}

func (c *controller) formatFullFromGroup(ctx context.Context, group *models.Group) (*service.Full, error) {
	groups, err := c.groupManager.GetByIDs(ctx, groupmanager.FormatIDsFromTraversalIDs(group.TraversalIDs))
	if err != nil {
		return nil, err
	}

	return service.GenerateFullFromGroups(groups), nil
}

func (c *controller) ofGroupModel(ctx context.Context, groups []*models.Group) ([]*Group, error) {
	var ofGroups = make([]*Group, 0)
	for _, item := range groups {
		fullEntity, err := c.formatFullFromGroup(ctx, item)
		if err != nil {
			return nil, err
		}
		ofGroups = append(ofGroups, &Group{
			ID:              item.ID,
			Name:            item.Name,
			Path:            item.Path,
			VisibilityLevel: item.VisibilityLevel,
			Description:     item.Description,
			ParentID:        item.ParentID,
			TraversalIDs:    item.TraversalIDs,
			UpdatedAt:       item.UpdatedAt,
			FullName:        fullEntity.FullName,
			FullPath:        fullEntity.FullPath,
		})
	}
	return ofGroups, nil
}

func (c *controller) UpdateRegionSelector(ctx context.Context, id uint, regionSelector RegionSelectors) error {
	// marshal struct to string
	regionSelectorBytes, err := yaml.Marshal(regionSelector)
	if err != nil {
		return herrors.NewErrUpdateFailed(herrors.GroupInDB, err.Error())
	}

	return c.groupManager.UpdateRegionSelector(ctx, id, string(regionSelectorBytes))
}
