package cluster

import (
	"context"

	herror "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

// Deprecated: Upgrade v1 to v2
func (c *controller) Upgrade(ctx context.Context, clusterID uint) error {
	const op = "cluster controller: upgrade to v2"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. validate infos
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return err
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return err
	}
	templateFromFile, err := c.clusterGitRepo.GetClusterTemplate(ctx, application.Name, cluster.Name)
	if err != nil {
		return err
	}

	// 2. match target template
	targetTemplate, ok := c.templateUpgradeMapper[templateFromFile.Name]
	if !ok {
		return perror.Wrapf(herror.ErrParamInvalid,
			"cluster template %s does not support upgrade", templateFromFile.Name)
	}
	targetRelease, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx,
		targetTemplate.Name, targetTemplate.Release)
	if err != nil {
		return err
	}

	// 3. sync gitops branch if restarts occur
	err = c.syncGitOpsBranch(ctx, application.Name, cluster.Name)
	if err != nil {
		return err
	}

	// 4. upgrade git repo files to v2
	_, err = c.clusterGitRepo.UpgradeCluster(ctx, &gitrepo.UpgradeValuesParam{
		Application:   application.Name,
		Cluster:       cluster.Name,
		Template:      templateFromFile,
		TargetRelease: targetRelease,
		BuildConfig:   &targetTemplate.BuildConfig,
	})
	if err != nil {
		return err
	}

	// 5. update template in db
	// TODO(zhuxu): remove strong dependencies on db updates, just print an err log when updates fail
	cluster.Template = targetRelease.TemplateName
	cluster.TemplateRelease = targetRelease.Name
	_, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return err
	}
	return nil
}
