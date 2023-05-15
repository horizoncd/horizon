package eventhandler

import (
	"context"

	eventhandlercfg "github.com/horizoncd/horizon/pkg/config/eventhandler"
	eventhandlersvc "github.com/horizoncd/horizon/pkg/eventhandler"
	"github.com/horizoncd/horizon/pkg/jobs"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
)

func New(ctx context.Context, eventHandlerConfig eventhandlercfg.Config,
	mgrs *managerparam.Manager) (jobs.Job, eventhandlersvc.Service) {
	// start event handler service to generate webhook log by events
	eventHandlerService := eventhandlersvc.NewService(ctx, mgrs, eventHandlerConfig)

	return func(ctx context.Context) {
		eventHandlerService.Start()
		<-ctx.Done()
		// graceful exit
		eventHandlerService.StopAndWait()
	}, eventHandlerService
}
