package application

import (
	"context"
)

var (
// Ctl Global instance of the application controller
// Ctl = NewController()
)

type Controller interface {
	// CreateApplication create an application
	CreateApplication(ctx context.Context, name string) error
}
