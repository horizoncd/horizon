package common

import (
	"context"

	herror "github.com/horizoncd/horizon/core/errors"
)

const contextPipelinerunIDKey = "pipelinerunID"

func ContextPipelinerunIDKey() string {
	return contextPipelinerunIDKey
}

func PipelinerunIDFromContext(ctx context.Context) (uint, error) {
	u, ok := ctx.Value(contextPipelinerunIDKey).(uint)
	if !ok {
		return 0, herror.ErrFailedToGetPipelinerunID
	}
	return u, nil
}
