package member

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/member/models"
)

var (
	// Mgr is the global member manager
	// Mgr = New()
)

type Manager interface {
	Create(ctx context.Context, member *models.Member) (models.Member, error)

	ListMember(ctx context.Context, query *q.Query) (int, []models.Member, error)

	GetByUserName(ctx context.Context, userName string) (models.Member, error)

	UpdateByID(ctx context.Context, id uint16, member *models.Member) (models.Member, error)
}


