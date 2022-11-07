package template

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/gitlab"
	hctx "g.hz.netease.com/horizon/pkg/context"
	perror "g.hz.netease.com/horizon/pkg/errors"
	gmanager "g.hz.netease.com/horizon/pkg/group/manager"
	groupModels "g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/group/service"
	membermanager "g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	tmanager "g.hz.netease.com/horizon/pkg/template/manager"
	"g.hz.netease.com/horizon/pkg/template/models"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/pkg/templaterelease/schema"
	"g.hz.netease.com/horizon/pkg/templaterepo"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	gitlabapi "github.com/xanzy/go-gitlab"
	"helm.sh/helm/v3/pkg/chart/loader"
)

const ChartNameFormat = "%s"
const GitlabName = "control"

type Controller interface {
	// ListTemplate list all template available
	ListTemplate(ctx context.Context) (Templates, error)
	// ListTemplateRelease list all releases of the specified template
	ListTemplateRelease(ctx context.Context, templateName string) (Releases, error)
	// GetTemplateSchema get schema for a template release
	GetTemplateSchema(ctx context.Context, releaseID uint, params map[string]string) (*Schemas, error)
	// ListTemplateByGroupID lists all template available by group ID
	ListTemplateByGroupID(ctx context.Context, groupID uint, withoutCI bool) (Templates, error)
	// ListTemplateReleaseByTemplateID lists all releases of the specified template
	ListTemplateReleaseByTemplateID(ctx context.Context, templateID uint) (Releases, error)
	// CreateTemplate creates a template with a release under a group
	CreateTemplate(ctx context.Context, groupID uint, request CreateTemplateRequest) (*Template, error)
	// CreateRelease downloads template archive and push it to harbor,
	// then creates a template release in database.
	CreateRelease(ctx context.Context, templateID uint, request CreateReleaseRequest) (*Release, error)
	// GetTemplate gets template by templateID
	GetTemplate(ctx context.Context, templateID uint) (*Template, error)
	// GetRelease gets release by releaseID
	GetRelease(ctx context.Context, releaseID uint) (*Release, error)
	// DeleteTemplate deletes a template by ID
	DeleteTemplate(ctx context.Context, templateID uint) error
	// DeleteRelease deletes a template release by ID
	DeleteRelease(ctx context.Context, releaseID uint) error
	// UpdateTemplate deletes a template by ID
	UpdateTemplate(ctx context.Context, templateID uint, request UpdateTemplateRequest) error
	// UpdateRelease deletes a template release by ID
	UpdateRelease(ctx context.Context, releaseID uint, request UpdateReleaseRequest) error
	// SyncReleaseToRepo downloads template from gitlab, packages the template and uploads it to chart repo
	SyncReleaseToRepo(ctx context.Context, releaseID uint) error
}

type controller struct {
	gitlabLib            gitlab.Interface
	templateRepo         templaterepo.TemplateRepo
	groupMgr             gmanager.Manager
	templateMgr          tmanager.Manager
	templateReleaseMgr   trmanager.Manager
	memberMgr            membermanager.Manager
	memberSvc            memberservice.Service
	templateSchemaGetter schema.Getter
}

var _ Controller = (*controller)(nil)

// NewController initializes a new controller
func NewController(param *param.Param, gitlabLib gitlab.Interface, repo templaterepo.TemplateRepo) Controller {
	return &controller{
		gitlabLib:            gitlabLib,
		templateMgr:          param.TemplateMgr,
		templateReleaseMgr:   param.TemplateReleaseManager,
		templateSchemaGetter: param.TemplateSchemaGetter,
		templateRepo:         repo,
		memberMgr:            param.MemberManager,
		memberSvc:            param.MemberService,
		groupMgr:             param.GroupManager,
	}
}

func (c *controller) ListTemplate(ctx context.Context) (Templates, error) {
	const op = "template controller: listTemplate"
	defer wlog.Start(ctx, op).StopPrint()

	var (
		tpls Templates
		err  error
	)

	if selfOnly, ok := ctx.Value(hctx.TemplateListSelfOnly).(bool); ok && selfOnly {
		tpls, err = c.listTemplateByUser(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		templateModels, err := c.templateMgr.List(ctx)
		if err != nil {
			return nil, err
		}

		for _, template := range templateModels {
			if c.checkHasOnlyOwnerPermissionForTemplate(ctx, template) {
				tpls = append(tpls, toTemplate(template))
			}
		}
	}

	if withFullpath, ok := ctx.Value(hctx.TemplateWithFullPath).(bool); ok && withFullpath {
		tpls, err = c.addFullPath(ctx, tpls)
		if err != nil {
			return nil, err
		}
	}

	return tpls, nil
}

// listTemplateByUser returns all templates a user obtaining, that means has owner permission
func (c *controller) listTemplateByUser(ctx context.Context) (Templates, error) {
	// get current user
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, perror.WithMessage(err, "no user in context")
	}

	// get groups authorized to current user
	groupIDs, err := c.memberMgr.ListResourceOfMemberInfoByRole(
		ctx, membermodels.TypeGroup, currentUser.GetID(), role.Owner)
	if err != nil {
		return nil, perror.WithMessage(err, "failed to list group resource of current user")
	}

	// get these groups' subGroups
	subGroups, err := c.groupMgr.GetSubGroupsByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, perror.WithMessage(err, "failed to get groups")
	}

	groupIDs = nil
	for _, group := range subGroups {
		var member *membermodels.Member
		if member, err = c.memberMgr.Get(ctx, membermodels.TypeGroup,
			group.ID, membermodels.MemberUser, currentUser.GetID()); err != nil {
			return nil, err
		}
		if member == nil || member.Role == role.Owner {
			groupIDs = append(groupIDs, group.ID)
		}
	}

	// list templates of these subGroups
	templates, err := c.templateMgr.ListByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	// get templates authorized to current user
	authorizedTemplateIDs, err := c.memberMgr.ListResourceOfMemberInfoByRole(ctx,
		membermodels.TypeTemplate, currentUser.GetID(), role.Owner)
	if err != nil {
		return nil, err
	}

	// all applicationIDs, including:
	// (1) templates under the authorized groups
	// (2) authorized templates directly
	authorizedTemplates, err := c.templateMgr.ListByIDs(ctx, authorizedTemplateIDs)
	if err != nil {
		return nil, err
	}

	set := make(map[uint]struct{})
	for _, template := range templates {
		set[template.ID] = struct{}{}
	}

	filter := func(t *models.Template) bool {
		if _, ok := set[t.ID]; !ok {
			set[t.ID] = struct{}{}
			return true
		}
		return false
	}

	for _, t := range authorizedTemplates {
		if filter(t) {
			templates = append(templates, t)
		}
	}

	return toTemplates(templates), nil
}

func (c *controller) addFullPath(ctx context.Context, tpls Templates) (Templates, error) {
	if withFullPath, ok := ctx.Value(hctx.TemplateWithFullPath).(bool); ok && withFullPath {
		for i, tpl := range tpls {
			if tpl.GroupID == service.RootGroupID {
				tpls[i].FullPath = "/" + tpl.Name
				continue
			}
			group, err := c.groupMgr.GetByID(ctx, tpl.GroupID)
			if err != nil {
				return nil, err
			}
			groupStrArr := strings.Split(group.TraversalIDs, ",")
			groupIDArr := make([]uint, 0, len(groupStrArr))
			for _, groupStr := range groupStrArr {
				t, err := strconv.Atoi(groupStr)
				if err != nil {
					return nil, perror.Wrap(herrors.ErrTemplateParamInvalid, err.Error())
				}
				groupIDArr = append(groupIDArr, uint(t))
			}
			groups, err := c.groupMgr.GetByIDs(ctx, groupIDArr)
			if err != nil {
				return nil, err
			}
			groupMap := make(map[uint]*groupModels.Group, len(groupIDArr))
			for _, group := range groups {
				groupMap[group.ID] = group
			}
			fullpath := strings.Builder{}
			for _, groupID := range groupIDArr {
				fullpath.WriteString("/")
				fullpath.WriteString(groupMap[groupID].Path)
			}
			fullpath.WriteString("/")
			fullpath.WriteString(tpl.Name)
			tpls[i].FullPath = fullpath.String()
		}
	}
	return tpls, nil
}

func (c *controller) ListTemplateRelease(ctx context.Context, templateName string) (_ Releases, err error) {
	const op = "template controller: listTemplateRelease"
	defer wlog.Start(ctx, op).StopPrint()

	template, err := c.templateMgr.GetByName(ctx, templateName)
	if err != nil {
		return nil, err
	}

	if !c.checkHasOnlyOwnerPermissionForTemplate(ctx, template) {
		return nil, perror.Wrapf(herrors.ErrForbidden,
			"you have no permission to access this resource\n"+
				"template name = %s", templateName)
	}

	templateReleaseModels, err := c.templateReleaseMgr.ListByTemplateName(ctx, templateName)
	if err != nil {
		return nil, err
	}

	templateReleaseModels = c.checkStatusForReleases(ctx, template, templateReleaseModels)

	for _, release := range templateReleaseModels {
		_ = c.templateReleaseMgr.UpdateByID(ctx, release.ID, release)
	}

	var releases Releases
	for _, release := range templateReleaseModels {
		if c.checkHasOnlyOwnerPermissionForRelease(ctx, release) {
			releases = append(releases, toRelease(release))
		}
	}
	sort.Sort(releases)
	return releases, nil
}

func (c *controller) GetTemplateSchema(ctx context.Context, releaseID uint,
	param map[string]string) (_ *Schemas, err error) {
	const op = "template controller: getTemplateSchema"
	defer wlog.Start(ctx, op).StopPrint()

	release, err := c.templateReleaseMgr.GetByID(ctx, releaseID)
	if err != nil {
		return nil, err
	}

	schemas, err := c.templateSchemaGetter.GetTemplateSchema(ctx, release.TemplateName, release.Name, param)
	if err != nil {
		return nil, err
	}

	return toSchemas(schemas), nil
}

// ListTemplateByGroupID lists all template available
func (c *controller) ListTemplateByGroupID(ctx context.Context, groupID uint, withoutCI bool) (Templates, error) {
	const op = "template controller: listTemplateByGroupID"
	defer wlog.Start(ctx, op).StopPrint()

	if !c.groupMgr.GroupExist(ctx, groupID) {
		reason := fmt.Sprintf("group not found: %d", groupID)
		return nil, perror.Wrap(herrors.NewErrNotFound(herrors.GroupInDB, reason), reason)
	}

	if listRecursively, ok := ctx.Value(hctx.TemplateListRecursively).(bool); ok && listRecursively {
		return c.listTemplateByGroupIDRecursively(ctx, groupID, withoutCI)
	}

	templates, err := c.templateMgr.ListByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var tpls Templates
	for _, template := range templates {
		if c.checkHasOnlyOwnerPermissionForTemplate(ctx, template) {
			if template.WithoutCI == withoutCI {
				tpls = append(tpls, toTemplate(template))
			}
		}
	}

	if withFullpath, ok := ctx.Value(hctx.TemplateWithFullPath).(bool); ok && withFullpath {
		tpls, err = c.addFullPath(ctx, tpls)
		if err != nil {
			return nil, err
		}
	}
	return tpls, err
}

func (c *controller) listTemplateByGroupIDRecursively(ctx context.Context,
	groupID uint, withoutCI bool) (Templates, error) {
	if !c.groupMgr.GroupExist(ctx, groupID) {
		reason := fmt.Sprintf("group not found: %d", groupID)
		return nil, perror.Wrap(herrors.NewErrNotFound(herrors.GroupInDB, reason), reason)
	}

	groupIDs := make([]uint, 0)

	if c.groupMgr.IsRootGroup(ctx, groupID) {
		return c.ListTemplate(ctx)
	}
	group, err := c.groupMgr.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	idStrs := strings.Split(group.TraversalIDs, ",")
	for _, idStr := range idStrs {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"failed to parse traversal ID:\n"+
					"traversal ID = %s", group.TraversalIDs)
		}
		groupIDs = append(groupIDs, uint(id))
	}

	templates, err := c.templateMgr.ListByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	var tpls Templates
	for _, template := range templates {
		if c.checkHasOnlyOwnerPermissionForTemplate(ctx, template) {
			if template.WithoutCI == withoutCI {
				tpls = append(tpls, toTemplate(template))
			}
		}
	}

	if withFullpath, ok := ctx.Value(hctx.TemplateWithFullPath).(bool); ok && withFullpath {
		tpls, err = c.addFullPath(ctx, tpls)
		if err != nil {
			return nil, err
		}
	}
	return tpls, err
}

// ListTemplateReleaseByTemplateID lists all releases of the specified template
func (c *controller) ListTemplateReleaseByTemplateID(ctx context.Context, templateID uint) (Releases, error) {
	const op = "template controller: listTemplateReleaseByTemplateID"
	defer wlog.Start(ctx, op).StopPrint()

	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	template, err := c.templateMgr.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if !c.checkHasOnlyOwnerPermissionForTemplate(ctx, template) {
		return nil, perror.Wrapf(herrors.ErrForbidden,
			"you have no permission to access this resource:\n"+
				"template id = %d", templateID)
	}

	releases, err := c.templateReleaseMgr.ListByTemplateID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	releases = c.checkStatusForReleases(ctx, template, releases)

	for _, release := range releases {
		_ = c.templateReleaseMgr.UpdateByID(ctx, release.ID, release)
	}

	if user.IsAdmin() {
		return toReleases(releases), nil
	}

	var releaseModels Releases
	for _, release := range releases {
		if c.checkHasOnlyOwnerPermissionForRelease(ctx, release) {
			releaseModels = append(releaseModels, toRelease(release))
		}
	}
	return releaseModels, nil
}

func (c *controller) CreateTemplate(ctx context.Context,
	groupID uint, request CreateTemplateRequest) (*Template, error) {
	const op = "template controller: createTemplate"
	defer wlog.Start(ctx, op).StopPrint()

	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// skip root group check
	if !c.groupMgr.GroupExist(ctx, groupID) {
		reason := fmt.Sprintf("group not found: %d", groupID)
		return nil, perror.Wrap(herrors.NewErrNotFound(herrors.GroupInDB, reason), reason)
	}

	// check if template exist
	if _, err = c.templateMgr.GetByName(ctx, request.Name); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return nil, err
		}
	} else {
		return nil, perror.Wrap(herrors.ErrNameConflict, "an template with the same name already exists, "+
			"please do not create it again")
	}

	template, err := request.toTemplateModel(ctx)
	if err != nil {
		return nil, err
	}
	template.ChartName = template.Name
	template.GroupID = groupID

	template, err = c.templateMgr.Create(ctx, template)
	if err != nil {
		return nil, err
	}

	_, err = c.memberMgr.Create(ctx, &membermodels.Member{
		ResourceType: membermodels.TypeTemplate,
		ResourceID:   template.ID,
		Role:         role.Owner,
		MemberType:   membermodels.MemberUser,
		MemberNameID: user.GetID(),
		GrantedBy:    user.GetID(),
		CreatedBy:    user.GetID(),
	})
	if err != nil {
		return nil, err
	}

	return toTemplate(template), nil
}

func (c *controller) CreateRelease(ctx context.Context,
	templateID uint, request CreateReleaseRequest) (*Release, error) {
	const op = "template controller: createTemplateRelease"
	defer wlog.Start(ctx, op).StopPrint()

	template, err := c.templateMgr.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	release, err := request.toReleaseModel(ctx, template)
	if err != nil {
		return nil, err
	}

	if syncToRepo, ok := ctx.Value(hctx.ReleaseSyncToRepo).(bool); !ok || (ok && syncToRepo) {
		tag, chartBytes, err := c.getTag(ctx, template.Repository, template.ChartName, release.Name)
		if err != nil {
			return nil, err
		}
		chartVersion := fmt.Sprintf(common.ChartVersionFormat, release.Name, tag.Commit.ShortID)
		err = c.syncReleaseToRepo(chartBytes, template.ChartName, chartVersion)
		if err != nil {
			return nil, err
		}
		release.CommitID = tag.Commit.ShortID
		release.SyncStatus = trmodels.StatusSucceed
		release.ChartVersion = chartVersion
	} else {
		release.SyncStatus = trmodels.StatusOutOfSync
	}

	release.Template = template.ID
	release.LastSyncAt = time.Now()

	var newRelease *trmodels.TemplateRelease
	if newRelease, err = c.templateReleaseMgr.Create(ctx, release); err != nil {
		return nil, err
	}
	return toRelease(newRelease), nil
}

// GetTemplate gets template by templateID
func (c *controller) GetTemplate(ctx context.Context, templateID uint) (*Template, error) {
	const op = "template controller: getTemplate"
	defer wlog.Start(ctx, op).StopPrint()

	template, err := c.templateMgr.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if !c.checkHasOnlyOwnerPermissionForTemplate(ctx, template) {
		return nil, perror.Wrapf(herrors.ErrForbidden,
			"you have no permission to access this resource:\n"+
				"template id = %d", templateID)
	}

	tpl := toTemplate(template)
	withRelease, ok := ctx.Value(hctx.TemplateWithRelease).(bool)
	if ok && withRelease {
		if templateReleases, err := c.templateReleaseMgr.
			ListByTemplateID(ctx, template.ID); err == nil {
			for _, release := range templateReleases {
				if c.checkHasOnlyOwnerPermissionForRelease(ctx, release) {
					tpl.Releases = append(tpl.Releases, toRelease(release))
				}
			}
		}
	}
	return tpl, nil
}

func (c *controller) GetRelease(ctx context.Context, releaseID uint) (*Release, error) {
	const op = "template controller: getRelease"
	defer wlog.Start(ctx, op).StopPrint()

	release, err := c.templateReleaseMgr.GetByID(ctx, releaseID)
	if err != nil {
		return nil, err
	}

	if !c.checkHasOnlyOwnerPermissionForRelease(ctx, release) {
		return nil, perror.Wrapf(herrors.ErrForbidden,
			"you have no permission to access this resource:\n"+
				"release id = %d", releaseID)
	}

	template, err := c.templateMgr.GetByID(ctx, release.Template)
	if err != nil {
		return nil, err
	}

	if !c.checkHasOnlyOwnerPermissionForTemplate(ctx, template) {
		return nil, perror.Wrapf(herrors.ErrForbidden,
			"you have no permission to access this resource:\n"+
				"template id = %d, release id = %d", release.Template, releaseID)
	}

	return toRelease(release), nil
}

func (c *controller) DeleteTemplate(ctx context.Context, templateID uint) error {
	const op = "template controller: deleteTemplate"
	defer wlog.Start(ctx, op).StopPrint()

	releases, err := c.templateReleaseMgr.ListByTemplateID(ctx, templateID)
	if err != nil {
		return err
	}
	if len(releases) != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist,
			"this template cannot be deleted because there are releases under this template.")
	}

	ctx = context.WithValue(ctx, hctx.TemplateOnlyRefCount, true)
	_, count, err := c.templateMgr.GetRefOfApplication(ctx, templateID)
	if err != nil {
		return err
	}
	if count != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist,
			"this template cannot be deleted because it was used by applications.")
	}

	_, count, err = c.templateMgr.GetRefOfCluster(ctx, templateID)
	if err != nil {
		return err
	}
	if count != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist, "this template cannot be deleted because it was used by clusters.")
	}

	return c.templateMgr.DeleteByID(ctx, templateID)
}

func (c *controller) DeleteRelease(ctx context.Context, releaseID uint) error {
	const op = "template controller: deleteRelease"
	defer wlog.Start(ctx, op).StopPrint()

	ctx = context.WithValue(ctx, hctx.TemplateOnlyRefCount, true)
	_, count, err := c.templateReleaseMgr.GetRefOfApplication(ctx, releaseID)
	if err != nil {
		return err
	}
	if count != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist, "this release cannot be deleted because it was used by applications.")
	}

	_, count, err = c.templateReleaseMgr.GetRefOfCluster(ctx, releaseID)
	if err != nil {
		return err
	}
	if count != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist, "this release cannot be deleted because it was used by clusters.")
	}

	return c.templateReleaseMgr.DeleteByID(ctx, releaseID)
}

// UpdateTemplate deletes a template by ID
func (c *controller) UpdateTemplate(ctx context.Context, templateID uint, request UpdateTemplateRequest) error {
	const op = "template controller: updateTemplate"
	defer wlog.Start(ctx, op).StopPrint()

	template, err := c.templateMgr.GetByID(ctx, templateID)
	if err != nil {
		return err
	}

	releases, err := c.templateReleaseMgr.ListByTemplateID(ctx, template.ID)
	if err != nil {
		return err
	}

	if len(releases) != 0 && request.Repository != "" &&
		request.Repository != template.Repository {
		return perror.Wrapf(herrors.ErrForbidden,
			"can not modify template repository while releases existing:\n"+
				"releases numbers: %d", len(releases))
	}

	tplUpdate, err := request.toTemplateModel(ctx)
	if err != nil {
		return err
	}

	return c.templateMgr.UpdateByID(ctx, templateID, tplUpdate)
}

// UpdateRelease deletes a template release by ID
func (c *controller) UpdateRelease(ctx context.Context, releaseID uint, request UpdateReleaseRequest) error {
	const op = "template controller: updateRelease"
	defer wlog.Start(ctx, op).StopPrint()

	trUpdate, err := request.toReleaseModel(ctx)
	if err != nil {
		return err
	}

	return c.templateReleaseMgr.UpdateByID(ctx, releaseID, trUpdate)
}

func (c *controller) SyncReleaseToRepo(ctx context.Context, releaseID uint) error {
	const op = "template controller: syncReleaseToRepo"
	defer wlog.Start(ctx, op).StopPrint()

	release, err := c.templateReleaseMgr.GetByID(ctx, releaseID)
	if err != nil {
		return err
	}

	template, err := c.templateMgr.GetByID(ctx, release.Template)
	if err != nil {
		return err
	}

	tag, chartBytes, err := c.getTag(ctx, template.Repository, template.ChartName, release.Name)
	if err != nil {
		return err
	}
	chartVersion := fmt.Sprintf(common.ChartVersionFormat, release.Name, tag.Commit.ShortID)
	err = c.syncReleaseToRepo(chartBytes, template.ChartName, chartVersion)
	if err != nil {
		_ = c.handleReleaseSyncStatus(ctx, release, tag.Commit.ShortID, err.Error())
	} else {
		_ = c.handleReleaseSyncStatus(ctx, release, tag.Commit.ShortID, "")
	}
	return err
}

func (c *controller) handleReleaseSyncStatus(ctx context.Context,
	release *trmodels.TemplateRelease, commitID string, failedReason string) error {
	if failedReason == "" {
		release.SyncStatus = trmodels.StatusSucceed
		release.FailedReason = ""
		release.ChartVersion = fmt.Sprintf(common.ChartVersionFormat, release.Name, commitID)
	} else {
		release.FailedReason = failedReason
		release.SyncStatus = trmodels.StatusFailed
	}
	release.CommitID = commitID
	release.LastSyncAt = time.Now()
	return c.templateReleaseMgr.UpdateByID(ctx, release.ID, release)
}

func (c *controller) getTag(ctx context.Context, repository,
	name, tag string) (*gitlabapi.Tag, []byte, error) {
	URL, err := url.Parse(repository)
	if err != nil || repository == "" {
		return nil, nil, perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("failed to parse gitlab url: %s", err))
	}

	pidPattern := regexp.MustCompile(`^/(.+?)(?:\.git)?$`)
	matches := pidPattern.FindStringSubmatch(URL.Path)
	if len(matches) != 2 {
		return nil, nil, perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("failed to parse gitlab url: %s", err))
	}

	pid := matches[1]

	chartBytes, err := c.gitlabLib.GetRepositoryArchive(ctx, pid, tag)
	if err != nil {
		return nil, nil, err
	}

	t, err := c.gitlabLib.GetTag(ctx, pid, tag)
	if err != nil {
		return nil, nil, err
	}
	return t, chartBytes, nil
}

func (c *controller) checkStatusForReleases(ctx context.Context,
	template *models.Template, releases []*trmodels.TemplateRelease) []*trmodels.TemplateRelease {
	var wg sync.WaitGroup

	for i, release := range releases {
		wg.Add(1)
		go func(index int, r *trmodels.TemplateRelease) {
			releases[index], _ = c.checkStatusForRelease(ctx, template, r)
			wg.Done()
		}(i, release)
	}
	wg.Wait()

	return releases
}

func (c *controller) checkStatusForRelease(ctx context.Context,
	template *models.Template, release *trmodels.TemplateRelease) (*trmodels.TemplateRelease, error) {
	if release.SyncStatus != trmodels.StatusSucceed {
		return release, nil
	}

	tag, _, err := c.getTag(ctx, template.Repository, release.ChartName, release.Name)
	if err != nil {
		release.SyncStatus = trmodels.StatusUnknown
		return release, err
	}
	if tag.Commit.ShortID != release.CommitID {
		release.SyncStatus = trmodels.StatusOutOfSync
	}
	return release, nil
}

func (c *controller) syncReleaseToRepo(chartBytes []byte, name, tag string) error {
	chart, err := loader.LoadArchive(bytes.NewReader(chartBytes))
	if err != nil {
		return perror.Wrap(herrors.ErrLoadChartArchive, fmt.Sprintf("failed to load archive: %v", err))
	}
	chart.Metadata.Version = tag
	chart.Metadata.Name = name

	return c.templateRepo.UploadChart(chart)
}

func (c *controller) checkHasOnlyOwnerPermissionForTemplate(ctx context.Context,
	template *models.Template) bool {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return false
	}
	if user.IsAdmin() {
		return true
	}

	if template == nil || template.OnlyOwner == nil {
		return false
	}

	if !*template.OnlyOwner {
		return true
	}

	member, err := c.memberSvc.GetMemberOfResource(ctx, common.ResourceTemplate, strconv.Itoa(int(template.ID)))
	if err != nil || member == nil {
		return false
	}
	if member.Role == role.Owner {
		return true
	}
	return false
}

func (c *controller) checkHasOnlyOwnerPermissionForRelease(ctx context.Context,
	release *trmodels.TemplateRelease) bool {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return false
	}
	if user.IsAdmin() {
		return true
	}

	if release == nil || release.OnlyOwner == nil {
		return false
	}

	if !*release.OnlyOwner {
		return true
	}

	member, err := c.memberSvc.GetMemberOfResource(ctx, common.ResourceTemplate, strconv.Itoa(int(release.ID)))
	if err != nil || member == nil {
		return false
	}
	if member.Role == role.Owner {
		return true
	}
	return false
}
