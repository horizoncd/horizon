package dao

import (
	"golang.org/x/net/context"
)

type DAO interface {
	Create(ctx context.Context)
	Get()
	DeleteByID()
	List()
}
