package hook

import "context"

type Hook interface {
	Push(ctx context.Context, hooks Event)
}
