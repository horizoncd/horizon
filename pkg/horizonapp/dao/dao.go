package dao

import (
	"golang.org/x/net/context"
)

type DAO interface {
	Create(ctx context.Context) // Create data.
	Get()                       // Get data.
	DeleteByID()                // Delete data by ID.
	List()                      // List all data.
}
