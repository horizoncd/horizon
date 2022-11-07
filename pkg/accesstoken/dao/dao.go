package dao

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/accesstoken/models"
	"g.hz.netease.com/horizon/pkg/oauth/generate"
	oauthmodels "g.hz.netease.com/horizon/pkg/oauth/models"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
)

type dao struct {
	db *gorm.DB
}

type DAO interface {
	ListAccessTokensOfResource(ctx context.Context, resourceType string, resourceID uint,
		query *q.Query) ([]*models.AccessToken, int, error)
	ListOwnAccessTokens(ctx context.Context, query *q.Query) ([]*models.AccessToken, int, error)
	GetAccessToken(ctx context.Context, id uint) (*oauthmodels.Token, error)
}

func NewDAO(db *gorm.DB) DAO {
	return &dao{
		db: db,
	}
}

func (d *dao) ListAccessTokensOfResource(ctx context.Context, resourceType string, resourceID uint,
	query *q.Query) ([]*models.AccessToken, int, error) {
	var (
		pageSize   = common.DefaultPageSize
		pageNumber = common.DefaultPageNumber
		tokens     []*models.AccessToken
		total      int64
	)
	if query != nil {
		if query.PageSize > 1 {
			pageSize = query.PageSize
		}
		if query.PageNumber > 0 {
			pageNumber = query.PageNumber
		}
	}
	limit := pageSize
	offset := (pageNumber - 1) * pageSize

	result := d.db.WithContext(ctx).Table("tb_token as t").
		Joins("join tb_user as u on t.user_or_robot_identity = u.id").
		Joins("join tb_member as m on u.id = m.membername_id").
		Where("u.user_type = ?", usermodels.UserTypeRobot).
		Where("m.resource_type = ?", resourceType).
		Where("m.resource_id = ?", resourceID).
		Select("t.*, m.role as role").Offset(offset).Limit(limit).Scan(&tokens).Offset(0).Limit(-1).Count(&total)
	return tokens, int(total), result.Error
}

func (d *dao) ListOwnAccessTokens(ctx context.Context, query *q.Query) ([]*models.AccessToken, int, error) {
	var (
		pageSize   = common.DefaultPageSize
		pageNumber = common.DefaultPageNumber
		tokens     []*models.AccessToken
		total      int64
	)
	if query != nil {
		if query.PageSize > 1 {
			pageSize = query.PageSize
		}
		if query.PageNumber > 0 {
			pageNumber = query.PageNumber
		}
	}
	limit := pageSize
	offset := (pageNumber - 1) * pageSize
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, int(total), err
	}

	result := d.db.WithContext(ctx).Table("tb_token").
		Where("user_id = ?", currentUser.GetID()).
		Where("code like ?", fmt.Sprintf("%s%%", generate.AccessTokenPrefix)).
		Offset(offset).Limit(limit).
		Find(&tokens).Offset(0).Limit(-1).Count(&total)
	return tokens, int(total), result.Error
}

func (d *dao) GetAccessToken(ctx context.Context, id uint) (*oauthmodels.Token, error) {
	var token oauthmodels.Token

	result := d.db.WithContext(ctx).Model(token).Where("id = ?", id).First(&token)
	return &token, result.Error
}
