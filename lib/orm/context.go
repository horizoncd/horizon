package orm

import (
	"context"

	he "g.hz.netease.com/horizon/core/errors"
	perrors "g.hz.netease.com/horizon/pkg/errors"

	"gorm.io/gorm"
)

const ormKey = "ORM"

func Key() string {
	return ormKey
}

// FromContext returns orm from context
func FromContext(ctx context.Context) (*gorm.DB, error) {
	o, ok := ctx.Value(ormKey).(*gorm.DB)
	if !ok {
		return nil, perrors.Wrap(he.ErrFailedToGetORM, "cannot get the ORM from context")
	}
	return o, nil
}

// NewContext returns new context with orm
func NewContext(ctx context.Context, o *gorm.DB) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ormKey, o) // nolint
}
