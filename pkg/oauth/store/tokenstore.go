package store

import (
	"g.hz.netease.com/horizon/pkg/oauth/models"
	"golang.org/x/net/context"
)

type DbTokenStore struct {
}

func (d *DbTokenStore) Create(ctx context.Context, token *models.Token) (*models.Token, error) {
	return nil, nil
}

func (d *DbTokenStore) DeleteByCode(ctx context.Context, code string) (*models.Token, error) {
	return nil, nil
}

func (d *DbTokenStore) DeleteByClientID(ctx context.Context, code string) error {
	return nil
}

func (d *DbTokenStore) Get(ctx context.Context, code string) (*models.Token, error) {
	return nil, nil
}
