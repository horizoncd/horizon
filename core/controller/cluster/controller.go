package cluster

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/deployer"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	groupsvc "g.hz.netease.com/horizon/pkg/group/service"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/jsonschema"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	CreateCluster(ctx context.Context, applicationID uint, environment, region string,
		request *CreateClusterRequest) (*GetClusterResponse, error)
}

type controller struct {
	deployer             deployer.Deployer
	applicationMgr       appmanager.Manager
	templateReleaseMgr   trmanager.Manager
	templateSchemaGetter templateschema.Getter
	envMgr               envmanager.Manager
	regionMgr            regionmanager.Manager
	groupSvc             groupsvc.Service
}

var _ Controller = (*controller)(nil)

func NewController(clusterGitRepo gitrepo.ClusterGitRepo, cd cd.CD,
	templateSchemaGetter templateschema.Getter) Controller {
	return &controller{
		deployer:             deployer.NewDeployer(clusterGitRepo, cd),
		applicationMgr:       appmanager.Mgr,
		templateReleaseMgr:   trmanager.Mgr,
		templateSchemaGetter: templateSchemaGetter,
		envMgr:               envmanager.Mgr,
		regionMgr:            regionmanager.Mgr,
		groupSvc:             groupsvc.Svc,
	}
}

func (c *controller) CreateCluster(ctx context.Context, applicationID uint,
	environment, region string, r *CreateClusterRequest) (_ *GetClusterResponse, err error) {
	const op = "cluster controller: create cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. get application
	application, err := c.applicationMgr.GetByID(ctx, applicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. validate
	if err := validateClusterName(application.Name, r.Name); err != nil {
		return nil, errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), err)
	}
	if err := c.validateTemplateInput(ctx,
		application.Template, application.TemplateRelease, r.TemplateInput); err != nil {
		return nil, errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), fmt.Sprintf("request body validate err: %v", err))
	}

	// 3. get environmentRegion
	er, err := c.envMgr.GetByEnvironmentAndRegion(ctx, environment, region)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 4. get regionEntity
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, region)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 5. get templateRelease
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, application.Template, application.TemplateRelease)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 6. create cluster, after created, params.Cluster is the newest cluster
	cluster := r.toClusterModel(application, er)
	cluster.CreatedBy = currentUser.GetID()
	cluster.UpdatedBy = currentUser.GetID()
	params := &deployer.Params{
		Environment:         environment,
		Application:         application,
		Cluster:             cluster,
		RegionEntity:        regionEntity,
		PipelineJSONBlob:    r.TemplateInput.Pipeline,
		ApplicationJSONBlob: r.TemplateInput.Application,
		TemplateRelease:     tr,
	}

	if err := c.deployer.CreateCluster(ctx, params); err != nil {
		return nil, errors.E(op, err)
	}

	// 7. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, params.Cluster.Name)

	return ofClusterModel(application, params.Cluster, er, fullPath,
		r.TemplateInput.Pipeline, r.TemplateInput.Application), nil
}

// validateTemplateInput validate templateInput is valid for template schema
func (c *controller) validateTemplateInput(ctx context.Context,
	template, release string, templateInput *TemplateInput) error {
	schema, err := c.templateSchemaGetter.GetTemplateSchema(ctx, template, release)
	if err != nil {
		return err
	}
	if err := jsonschema.Validate(schema.Application.JSONSchema, templateInput.Application); err != nil {
		return err
	}
	return jsonschema.Validate(schema.Pipeline.JSONSchema, templateInput.Pipeline)
}

// validateClusterName validate cluster name
// 1. name length must be less than 53
// 2. name must match pattern ^(([a-z][-a-z0-9]*)?[a-z0-9])?$
// 3. name must start with application name
func validateClusterName(applicationName, name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 53 {
		return fmt.Errorf("name must not exceed 53 characters")
	}

	// cannot start with a digit.
	if name[0] >= '0' && name[0] <= '9' {
		return fmt.Errorf("name cannot start with a digit")
	}

	if !strings.HasPrefix(name, applicationName) {
		return fmt.Errorf("cluster name must start with application name")
	}

	pattern := `^(([a-z][-a-z0-9]*)?[a-z0-9])?$`
	r := regexp.MustCompile(pattern)
	if !r.MatchString(name) {
		return fmt.Errorf("invalid cluster name, regex used for validation is %v", pattern)
	}

	return nil
}
