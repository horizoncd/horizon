package orm

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

const ormKey = "ORM"

func ORMKey() string {
	return ormKey
}

// FromContext returns orm from context
func FromContext(ctx context.Context) (*gorm.DB, error) {
	o, ok := ctx.Value(ormKey).(*gorm.DB)
	if !ok {
		return nil, errors.New("cannot get the ORM from context")
	}
	return o, nil
}
