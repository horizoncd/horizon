package cluster

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/horizoncd/horizon/pkg/core/common"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/tekton"
	"github.com/horizoncd/horizon/pkg/git"
	prmodels "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	tokensvc "github.com/horizoncd/horizon/pkg/token/service"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"

	"github.com/mozillazg/go-pinyin"
)

func (c *controller) BuildDeploy(ctx context.Context, clusterID uint,
	r *BuildDeployRequest) (_ *BuildDeployResponse, err error) {
	const op = "cluster controller: build deploy"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	var gitRef, gitRefType = cluster.GitRef, cluster.GitRefType
	if r.Git != nil {
		if r.Git.Commit != "" {
			gitRefType = codemodels.GitRefTypeCommit
			gitRef = r.Git.Commit
		} else if r.Git.Tag != "" {
			gitRefType = codemodels.GitRefTypeTag
			gitRef = r.Git.Tag
		} else if r.Git.Branch != "" {
			gitRefType = codemodels.GitRefTypeBranch
			gitRef = r.Git.Branch
		}
	}

	commit, err := c.commitGetter.GetCommit(ctx, cluster.GitURL, gitRefType, gitRef)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	// 1. assemble artifact imageURL
	imageURL := assembleImageURL(regionEntity, application.Name, cluster.Name, gitRef, commit.ID)

	configCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}

	// 2. add pipelinerun in db
	pr := &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionBuildDeploy,
		Status:           string(prmodels.StatusCreated),
		Title:            r.Title,
		Description:      r.Description,
		GitURL:           cluster.GitURL,
		GitRefType:       gitRefType,
		GitRef:           gitRef,
		GitCommit:        commit.ID,
		ImageURL:         imageURL,
		LastConfigCommit: configCommit.Master,
		ConfigCommit:     configCommit.Gitops,
	}
	prCreated, err := c.pipelinerunMgr.Create(ctx, pr)
	if err != nil {
		return nil, err
	}

	// 3. generate a JWT token for tekton callback
	token, err := c.tokenSvc.CreateJWTToken(strconv.Itoa(int(currentUser.GetID())),
		c.tokenConfig.CallbackTokenExpireIn, tokensvc.WithPipelinerunID(prCreated.ID))
	if err != nil {
		return nil, err
	}

	// 4. create pipelinerun in k8s
	tektonClient, err := c.tektonFty.GetTekton(cluster.EnvironmentName)
	if err != nil {
		return nil, err
	}

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	clusterFiles, err := c.clusterGitRepo.GetCluster(ctx,
		application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}

	prGit := tekton.PipelineRunGit{
		URL:       cluster.GitURL,
		Subfolder: cluster.GitSubfolder,
		Commit:    commit.ID,
	}
	switch prCreated.GitRefType {
	case codemodels.GitRefTypeTag:
		prGit.Tag = prCreated.GitRef
	case codemodels.GitRefTypeBranch:
		prGit.Branch = prCreated.GitRef
	}

	ciEventID, err := tektonClient.CreatePipelineRun(ctx, &tekton.PipelineRun{
		Application:      application.Name,
		ApplicationID:    application.ID,
		Cluster:          cluster.Name,
		ClusterID:        cluster.ID,
		Environment:      cluster.EnvironmentName,
		Git:              prGit,
		ImageURL:         imageURL,
		Operator:         currentUser.GetEmail(),
		PipelinerunID:    prCreated.ID,
		PipelineJSONBlob: clusterFiles.PipelineJSONBlob,
		Region:           cluster.RegionName,
		RegionID:         regionEntity.ID,
		Template:         cluster.Template,
		Token:            token,
	})
	if err != nil {
		return nil, err
	}

	// update event id returned from tekton-trigger EventListener
	log.Infof(ctx, "received event id: %s from tekton-trigger EventListener, pipelinerunID: %d", ciEventID, pr.ID)
	err = c.pipelinerunMgr.UpdateCIEventIDByID(ctx, pr.ID, ciEventID)
	if err != nil {
		return nil, err
	}

	return &BuildDeployResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func assembleImageURL(regionEntity *regionmodels.RegionEntity,
	application, cluster, branch, commit string) string {
	// domain is harbor server
	domain := strings.TrimPrefix(regionEntity.Registry.Server, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	// time now
	timeFormat := "20060102150405"
	timeStr := time.Now().Format(timeFormat)

	// normalize branch
	args := pinyin.Args{
		Fallback: func(r rune, a pinyin.Args) []string {
			return []string{string(r)}
		},
	}
	normalizedBranch := strings.Join(pinyin.LazyPinyin(branch, args), "")
	normalizedBranch = regexp.MustCompile(`[^a-zA-Z0-9_.-]`).ReplaceAllString(normalizedBranch, "_")

	return path.Join(domain, regionEntity.Registry.Path, application,
		fmt.Sprintf("%v:%v-%v-%v", cluster, normalizedBranch, commit[:8], timeStr))
}

func (c *controller) GetDiff(ctx context.Context, clusterID uint, refType, ref string) (_ *GetDiffResponse, err error) {
	const op = "cluster controller: get diff"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get cluster
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 2. get application
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 3. get code commit
	var commit *git.Commit
	if ref != "" {
		commit, err = c.commitGetter.GetCommit(ctx, cluster.GitURL, refType, ref)
		if err != nil {
			return nil, err
		}
	}

	// 4.  get config diff
	diff, err := c.clusterGitRepo.CompareConfig(ctx, application.Name, cluster.Name, nil, nil)
	if err != nil {
		return nil, err
	}
	return c.ofClusterDiff(cluster.GitURL, refType, ref, commit, diff)
}

func (c *controller) ofClusterDiff(gitURL, refType, ref string, commit *git.Commit, diff string) (
	*GetDiffResponse, error) {
	var codeInfo *CodeInfo

	if commit != nil {
		historyLink, err := c.commitGetter.GetCommitHistoryLink(gitURL, ref)
		if err != nil {
			return nil, err
		}
		codeInfo = &CodeInfo{
			CommitID:  commit.ID,
			CommitMsg: commit.Message,
			Link:      historyLink,
		}
		switch refType {
		case codemodels.GitRefTypeTag:
			codeInfo.Tag = ref
		case codemodels.GitRefTypeBranch:
			codeInfo.Branch = ref
		}
	}

	return &GetDiffResponse{
		CodeInfo:   codeInfo,
		ConfigDiff: diff,
	}, nil
}
