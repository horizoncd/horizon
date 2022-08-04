package template

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"regexp"
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

const ChartNameFormat = "%d-%s"
const GitlabName = "control"

type Controller interface {
	// ListTemplate list all template available
	ListTemplate(ctx context.Context) (Templates, error)
	// ListTemplateRelease list all releases of the specified template
	ListTemplateRelease(ctx context.Context, templateName string) (Releases, error)
	// GetTemplateSchema get schema for a template release
	GetTemplateSchema(ctx context.Context, releaseID uint, params map[string]string) (*Schemas, error)
	// ListTemplateByGroupID lists all template available by group ID
	ListTemplateByGroupID(ctx context.Context, groupID uint) (Templates, error)
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
		groupMgr:             param.GroupManager,
	}
}

func (c *controller) ListTemplate(ctx context.Context) (_ Templates, err error) {
	const op = "template controller: listTemplate"
	defer wlog.Start(ctx, op).StopPrint()

	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	templateModels, err := c.templateMgr.List(ctx)
	if err != nil {
		return nil, err
	}

	var tpls Templates
	if user.IsAdmin() {
		tpls = toTemplates(templateModels)
	} else {
		var filteredModels []*models.Template
		for _, template := range templateModels {
			if template.OnlyAdmin != nil && !*template.OnlyAdmin {
				filteredModels = append(filteredModels, template)
			}
		}
		tpls = toTemplates(filteredModels)
	}

	if withFullpath, ok := ctx.Value(hctx.TemplateWithFullPath).(bool); ok && withFullpath {
		tpls, err = c.addFullPath(ctx, tpls)
		if err != nil {
			return nil, err
		}
	}

	return tpls, nil
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
				fullpath.WriteString(groupMap[groupID].Name)
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

	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	template, err := c.templateMgr.GetByName(ctx, templateName)
	if err != nil {
		return nil, err
	}

	if template.OnlyAdmin != nil &&
		!(!*template.OnlyAdmin && user.IsAdmin()) {
		return make(Releases, 0), nil
	}

	templateReleaseModels, err := c.templateReleaseMgr.ListByTemplateName(ctx, templateName)
	if err != nil {
		return nil, err
	}

	templateReleaseModels = c.checkStatusForReleases(ctx, template, templateReleaseModels)

	for _, release := range templateReleaseModels {
		_ = c.templateReleaseMgr.UpdateByID(ctx, release.ID, release)
	}

	if user.IsAdmin() {
		return toReleases(templateReleaseModels), nil
	}

	var filteredModels []*trmodels.TemplateRelease
	for _, release := range templateReleaseModels {
		if release.OnlyAdmin != nil && !*release.OnlyAdmin {
			filteredModels = append(filteredModels, release)
		}
	}
	return toReleases(filteredModels), nil
}

func (c *controller) GetTemplateSchema(ctx context.Context, releaseID uint,
	param map[string]string) (_ *Schemas, err error) {
	const op = "template controller: getTemplateSchema"
	defer wlog.Start(ctx, op).StopPrint()

	if err := checkPermission(ctx, c, common.ResourceTemplateRelease, releaseID); err != nil {
		return nil, err
	}

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
func (c *controller) ListTemplateByGroupID(ctx context.Context, groupID uint) (Templates, error) {
	const op = "template controller: listTemplateByGroupID"
	defer wlog.Start(ctx, op).StopPrint()

	user, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !c.groupMgr.GroupExist(ctx, groupID) {
		reason := fmt.Sprintf("group not found: %d", groupID)
		return nil, perror.Wrap(herrors.NewErrNotFound(herrors.GroupInDB, reason), reason)
	}

	templates, err := c.templateMgr.ListByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var tpls Templates
	if user.IsAdmin() {
		tpls = toTemplates(templates)
	} else {
		var filteredModels []*models.Template
		for _, template := range templates {
			if template.OnlyAdmin != nil && !*template.OnlyAdmin {
				filteredModels = append(filteredModels, template)
			}
		}
		tpls = toTemplates(filteredModels)
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

	var filteredModels []*trmodels.TemplateRelease
	for _, release := range releases {
		if release.OnlyAdmin != nil && !*release.OnlyAdmin {
			filteredModels = append(filteredModels, release)
		}
	}
	return toReleases(filteredModels), nil
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

	template, err := request.toTemplateModel(ctx)
	if err != nil {
		return nil, err
	}
	template.ChartName = fmt.Sprintf(ChartNameFormat, groupID, template.Name)
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
		_ = c.templateRepo.DeleteChart(template.ChartName, release.ChartVersion)
		return nil, err
	}
	return toRelease(newRelease), nil
}

// GetTemplate gets template by templateID
func (c *controller) GetTemplate(ctx context.Context, templateID uint) (*Template, error) {
	const op = "template controller: getTemplate"
	defer wlog.Start(ctx, op).StopPrint()

	if err := checkPermission(ctx, c, common.ResourceTemplate, templateID); err != nil {
		return nil, err
	}

	template, err := c.templateMgr.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}
	tpl := toTemplate(template)

	withRelease, ok := ctx.Value(hctx.TemplateWithRelease).(bool)
	if ok && withRelease {
		if templateReleases, err := c.templateReleaseMgr.
			ListByTemplateName(ctx, template.Name); err != nil {
			releases := toReleases(templateReleases)
			tpl.Releases = releases
		}
	}
	return tpl, nil
}

func (c *controller) GetRelease(ctx context.Context, releaseID uint) (*Release, error) {
	const op = "template controller: getRelease"
	defer wlog.Start(ctx, op).StopPrint()

	if err := checkPermission(ctx, c, common.ResourceTemplateRelease, releaseID); err != nil {
		return nil, err
	}

	templateRelease, err := c.templateReleaseMgr.GetByID(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	return toRelease(templateRelease), nil
}

func (c *controller) DeleteTemplate(ctx context.Context, templateID uint) error {
	const op = "template controller: deleteTemplate"
	defer wlog.Start(ctx, op).StopPrint()

	if err := checkPermission(ctx, c, common.ResourceTemplate, templateID); err != nil {
		return err
	}

	releases, err := c.templateReleaseMgr.ListByTemplateID(ctx, templateID)
	if err != nil {
		return err
	}
	if len(releases) != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist, "template still has release")
	}
	ctx = context.WithValue(ctx, hctx.TemplateOnlyRefCount, true)
	_, count, err := c.templateMgr.GetRefOfApplication(ctx, templateID)
	if err != nil {
		return err
	}
	if count != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist, "template has been used by application")
	}

	_, count, err = c.templateMgr.GetRefOfCluster(ctx, templateID)
	if err != nil {
		return err
	}
	if count != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist, "template has been used by cluster")
	}

	return c.templateMgr.DeleteByID(ctx, templateID)
}

func (c *controller) DeleteRelease(ctx context.Context, releaseID uint) error {
	const op = "template controller: deleteRelease"
	defer wlog.Start(ctx, op).StopPrint()

	if err := checkPermission(ctx, c, common.ResourceTemplateRelease, releaseID); err != nil {
		return err
	}

	ctx = context.WithValue(ctx, hctx.TemplateOnlyRefCount, true)
	_, count, err := c.templateReleaseMgr.GetRefOfApplication(ctx, releaseID)
	if err != nil {
		return err
	}
	if count != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist, "release has been used by application")
	}

	_, count, err = c.templateReleaseMgr.GetRefOfCluster(ctx, releaseID)
	if err != nil {
		return err
	}
	if count != 0 {
		return perror.Wrap(herrors.ErrSubResourceExist, "release template has been used by cluster")
	}

	release, err := c.templateReleaseMgr.GetByID(ctx, releaseID)
	if err != nil {
		return err
	}

	template, err := c.templateMgr.GetByID(ctx, release.Template)
	if err != nil {
		return err
	}

	if err := c.templateRepo.DeleteChart(template.ChartName, release.ChartVersion); err != nil {
		return err
	}

	return c.templateReleaseMgr.DeleteByID(ctx, releaseID)
}

// UpdateTemplate deletes a template by ID
func (c *controller) UpdateTemplate(ctx context.Context, templateID uint, request UpdateTemplateRequest) error {
	const op = "template controller: updateTemplate"
	defer wlog.Start(ctx, op).StopPrint()

	if err := checkPermission(ctx, c, common.ResourceTemplate, templateID); err != nil {
		return err
	}

	tplUpdate, err := request.toTemplateModel(ctx)
	if err != nil {
		return err
	}

	if _, err := c.templateMgr.GetByID(ctx, templateID); err != nil {
		return err
	}

	return c.templateMgr.UpdateByID(ctx, templateID, tplUpdate)
}

// UpdateRelease deletes a template release by ID
func (c *controller) UpdateRelease(ctx context.Context, releaseID uint, request UpdateReleaseRequest) error {
	const op = "template controller: updateRelease"
	defer wlog.Start(ctx, op).StopPrint()

	if err := checkPermission(ctx, c, common.ResourceTemplateRelease, releaseID); err != nil {
		return err
	}

	trUpdate, err := request.toReleaseModel(ctx)
	if err != nil {
		return err
	}

	if _, err := c.templateReleaseMgr.GetByID(ctx, releaseID); err != nil {
		return err
	}

	return c.templateReleaseMgr.UpdateByID(ctx, releaseID, trUpdate)
}

func (c *controller) SyncReleaseToRepo(ctx context.Context, releaseID uint) error {
	const op = "template controller: syncReleaseToRepo"
	defer wlog.Start(ctx, op).StopPrint()

	if err := checkPermission(ctx, c, common.ResourceTemplateRelease, releaseID); err != nil {
		return err
	}

	release, err := c.templateReleaseMgr.GetByID(ctx, releaseID)
	if err != nil {
		return err
	}

	if release.SyncStatus == trmodels.StatusSucceed {
		return nil
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

func checkPermission(ctx context.Context, c *controller, resource string, id uint) error {
	user, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	if user.IsAdmin() {
		return nil
	}

	switch resource {
	case common.ResourceTemplate:
		template, err := c.templateMgr.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if template.OnlyAdmin != nil && *template.OnlyAdmin {
			return perror.Wrap(herrors.ErrForbidden, "you can not access it")
		}
	case common.ResourceTemplateRelease:
		release, err := c.templateReleaseMgr.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if release.OnlyAdmin != nil && *release.OnlyAdmin {
			return perror.Wrap(herrors.ErrForbidden, "you can not access it")
		}
	}
	return nil
}
