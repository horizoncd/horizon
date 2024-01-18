package service

import (
	"context"

	"github.com/horizoncd/horizon/core/common"
	amodels "github.com/horizoncd/horizon/pkg/application/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	cmodels "github.com/horizoncd/horizon/pkg/cluster/models"
	gmodels "github.com/horizoncd/horizon/pkg/group/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/pr/models"
	"github.com/horizoncd/horizon/pkg/util/log"
)

type Service interface {
	OfPipelineBasic(ctx context.Context, pr,
		firstCanRollbackPipelinerun *models.Pipelinerun) (*models.PipelineBasic, error)
	OfPipelineBasics(ctx context.Context, prs []*models.Pipelinerun,
		firstCanRollbackPipelinerun *models.Pipelinerun) ([]*models.PipelineBasic, error)
	GetCheckByResource(ctx context.Context, resourceID uint,
		resourceType string) ([]*models.Check, error)
	// CreateUserMessage creates a user message on pipeline run
	CreateUserMessage(ctx context.Context, prID uint, content string) (*models.PRMessage, error)
	// CreateSystemMessage creates a system message on pipeline run
	CreateSystemMessage(ctx context.Context, prID uint, content string)
}

type service struct {
	manager *managerparam.Manager
}

var _ Service = (*service)(nil)

func NewService(manager *managerparam.Manager) Service {
	return &service{
		manager,
	}
}

func (s *service) OfPipelineBasic(ctx context.Context,
	pr, firstCanRollbackPipelinerun *models.Pipelinerun) (*models.PipelineBasic, error) {
	user, err := s.manager.UserMgr.GetUserByID(ctx, pr.CreatedBy)
	if err != nil {
		return nil, err
	}

	canRollback := func() bool {
		// set the firstCanRollbackPipelinerun that cannot rollback
		if firstCanRollbackPipelinerun != nil && pr.ID == firstCanRollbackPipelinerun.ID {
			return false
		}
		return pr.Action != models.ActionRestart && pr.Status == string(models.StatusOK)
	}()

	prBasic := &models.PipelineBasic{
		ID:               pr.ID,
		Title:            pr.Title,
		Description:      pr.Description,
		Action:           pr.Action,
		Status:           pr.Status,
		GitURL:           pr.GitURL,
		GitCommit:        pr.GitCommit,
		ImageURL:         pr.ImageURL,
		LastConfigCommit: pr.LastConfigCommit,
		ConfigCommit:     pr.ConfigCommit,
		CreatedAt:        pr.CreatedAt,
		UpdatedAt:        pr.UpdatedAt,
		StartedAt:        pr.StartedAt,
		FinishedAt:       pr.FinishedAt,
		CanRollback:      canRollback,
		CreatedBy: models.UserInfo{
			UserID:   pr.CreatedBy,
			UserName: user.Name,
		},
	}
	switch pr.GitRefType {
	case codemodels.GitRefTypeTag:
		prBasic.GitTag = pr.GitRef
	case codemodels.GitRefTypeBranch:
		prBasic.GitBranch = pr.GitRef
	}
	return prBasic, nil
}

func (s *service) OfPipelineBasics(ctx context.Context, prs []*models.Pipelinerun,
	firstCanRollbackPipelinerun *models.Pipelinerun) ([]*models.PipelineBasic, error) {
	pipelineBasics := make([]*models.PipelineBasic, 0, len(prs))
	for _, pr := range prs {
		pipelineBasic, err := s.OfPipelineBasic(ctx, pr, firstCanRollbackPipelinerun)
		if err != nil {
			return nil, err
		}
		pipelineBasics = append(pipelineBasics, pipelineBasic)
	}
	return pipelineBasics, nil
}
func (s *service) GetCheckByResource(ctx context.Context, resourceID uint,
	resourceType string) ([]*models.Check, error) {
	var (
		id      = resourceID
		app     *amodels.Application
		cluster *cmodels.Cluster
		group   *gmodels.Group
		err     error
	)

	resources := make([]common.Resource, 0)

	switch resourceType {
	case common.ResourceCluster:
		cluster, err = s.manager.ClusterMgr.GetByID(ctx, resourceID)
		if err != nil {
			return nil, err
		}
		id = cluster.ApplicationID
		resources = append(resources, common.Resource{
			ResourceID: cluster.ID,
			Type:       common.ResourceCluster,
		})
		fallthrough
	case common.ResourceApplication:
		app, err = s.manager.ApplicationMgr.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		id = app.GroupID
		resources = append(resources, common.Resource{
			ResourceID: app.ID,
			Type:       common.ResourceApplication,
		})
		fallthrough
	case common.ResourceGroup:
		group, err = s.manager.GroupMgr.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	ids, err := common.UnmarshalTraversalIDS(group.TraversalIDs)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		resources = append(resources, common.Resource{
			ResourceID: id,
			Type:       common.ResourceGroup,
		})
	}
	resources = append(resources, common.Resource{
		ResourceID: 0,
		Type:       common.ResourceGroup,
	})

	return s.manager.PRMgr.Check.GetByResource(ctx, resources...)
}

func (s *service) CreateUserMessage(ctx context.Context, prID uint,
	content string) (*models.PRMessage, error) {
	return s.createMessage(ctx, prID, content, false)
}

func (s *service) CreateSystemMessage(ctx context.Context, prID uint,
	content string) {
	_, err := s.createMessage(ctx, prID, content, true)
	if err != nil {
		log.Warningf(ctx, "failed to create system message: %v", err.Error())
	}
}

func (s *service) createMessage(ctx context.Context, prID uint, content string,
	system bool) (*models.PRMessage, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.manager.PRMgr.Message.Create(ctx, &models.PRMessage{
		PipelineRunID: prID,
		Content:       content,
		System:        system,
		CreatedBy:     currentUser.GetID(),
		UpdatedBy:     currentUser.GetID(),
	})
}
