package cluster

import (
	"context"
	"fmt"
	"strconv"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/cluster/cd"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	perror "github.com/horizoncd/horizon/pkg/errors"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	tokenservice "github.com/horizoncd/horizon/pkg/token/service"
	usermodel "github.com/horizoncd/horizon/pkg/user/models"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

func (c *controller) InternalDeployV2(ctx context.Context, clusterID uint,
	r *InternalDeployRequestV2) (_ *InternalDeployResponseV2, err error) {
	const op = "cluster controller: internal deploy v2"
	defer wlog.Start(ctx, op).StopPrint()

	// auth jwt token
	claims, user, err := c.retrieveClaimsAndUser(ctx)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrTokenInvalid, "%v", err.Error())
	}
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:     user.Name,
		FullName: user.FullName,
		ID:       user.ID,
		Email:    user.Email,
		Admin:    user.Admin,
	})

	// auth prID
	if claims.PipelinerunID == nil ||
		*claims.PipelinerunID != r.PipelinerunID {
		return nil, perror.Wrapf(herrors.ErrForbidden,
			"no permission to deploy with pipelineID = %v", r.PipelinerunID)
	}

	// 1. get pr, and do some validate
	pr, err := c.pipelinerunMgr.GetByID(ctx, r.PipelinerunID)
	if err != nil {
		return nil, err
	}
	if pr == nil || pr.ClusterID != clusterID {
		return nil, herrors.NewErrNotFound(herrors.Pipelinerun,
			fmt.Sprintf("cannot find the pipelinerun with id: %v", r.PipelinerunID))
	}

	// 2. get some relevant models
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 3. update image in git repo
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	log.Infof(ctx, "pipeline %v output content: %+v", r.PipelinerunID, r.Output)
	commit, err := c.clusterGitRepo.UpdatePipelineOutput(ctx, application.Name, cluster.Name,
		tr.ChartName, r.Output)
	if err != nil {
		return nil, perror.WithMessage(err, op)
	}

	// 4. update config commit and status
	if err := c.pipelinerunMgr.UpdateConfigCommitByID(ctx, pr.ID, commit); err != nil {
		return nil, err
	}
	updatePRStatus := func(pState prmodels.PipelineStatus, revision string) error {
		if err = c.pipelinerunMgr.UpdateStatusByID(ctx, pr.ID, pState); err != nil {
			log.Errorf(ctx, "UpdateStatusByID error, pr = %d, status = %s, err = %v",
				pr.ID, pState, err)
			return err
		}
		log.Infof(ctx, "InternalDeploy status, pr = %d, status = %s, revision = %s",
			pr.ID, pState, revision)
		return nil
	}
	if err := updatePRStatus(prmodels.StatusCommitted, commit); err != nil {
		return nil, err
	}

	// 5. merge branch from gitops to master  and update status
	masterRevision, err := c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name,
		gitrepo.GitOpsBranch, c.clusterGitRepo.DefaultBranch(), &pr.ID)
	if err != nil {
		return nil, perror.WithMessage(err, op)
	}
	if err := updatePRStatus(prmodels.StatusMerged, masterRevision); err != nil {
		return nil, err
	}

	// 6. create cluster in cd system
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}
	repoInfo := c.clusterGitRepo.GetRepoInfo(ctx, application.Name, cluster.Name)
	if err := c.cd.CreateCluster(ctx, &cd.CreateClusterParams{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		GitRepoURL:   repoInfo.GitRepoURL,
		ValueFiles:   repoInfo.ValueFiles,
		RegionEntity: regionEntity,
		Namespace:    envValue.Namespace,
	}); err != nil {
		return nil, err
	}

	// 7. reset cluster status
	if cluster.Status == common.ClusterStatusFreed {
		cluster.Status = common.ClusterStatusEmpty
		cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
		if err != nil {
			return nil, err
		}
	}

	// 8. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    masterRevision,
	}); err != nil {
		return nil, err
	}

	// 9. update status
	if err := updatePRStatus(prmodels.StatusOK, masterRevision); err != nil {
		return nil, err
	}

	// 10. record event
	if _, err := c.eventMgr.CreateEvent(ctx, &eventmodels.Event{
		EventSummary: eventmodels.EventSummary{
			ResourceType: common.ResourceCluster,
			EventType:    eventmodels.ClusterBuildDeployed,
			ResourceID:   cluster.ID,
		},
	}); err != nil {
		log.Warningf(ctx, "failed to create event, err: %s", err.Error())
	}

	return &InternalDeployResponseV2{
		PipelinerunID: pr.ID,
		Commit:        commit,
	}, nil
}

func (c *controller) retrieveClaimsAndUser(ctx context.Context) (*tokenservice.Claims, *usermodel.User, error) {
	jwtTokenString, err := common.JWTTokenStringFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	claims, err := c.tokenSvc.ParseJWTToken(jwtTokenString)
	if err != nil {
		return nil, nil, err
	}
	userID, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return nil, nil, err
	}
	user, err := c.userManager.GetUserByID(ctx, uint(userID))
	if err != nil {
		return nil, nil, err
	}
	return &claims, user, nil
}

func (c *controller) InternalGetClusterStatus(ctx context.Context,
	clusterID uint) (_ *GetClusterStatusResponse, err error) {
	// auth jwt token
	_, user, err := c.retrieveClaimsAndUser(ctx)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrTokenInvalid, "%v", err.Error())
	}
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:     user.Name,
		FullName: user.FullName,
		ID:       user.ID,
		Email:    user.Email,
		Admin:    user.Admin,
	})
	return c.GetClusterStatus(ctx, clusterID)
}
