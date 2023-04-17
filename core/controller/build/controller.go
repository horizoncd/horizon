package build

import (
	"golang.org/x/net/context"
)

// Controller get build schema.
type Controller interface {
	GetSchema(ctx context.Context) (*Schema, error)
}

type controller struct {
	schema *Schema
}

func NewController(schema *Schema) Controller {
	return &controller{schema: schema}
}

func (c controller) GetSchema(_ context.Context) (*Schema, error) {
	return c.schema, nil
}

var _ Controller = (*controller)(nil)
