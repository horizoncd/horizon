package jobwebhook

import (
	"context"
	"log"

	eventhandlercfg "github.com/horizoncd/horizon/pkg/config/eventhandler"
	webhookcfg "github.com/horizoncd/horizon/pkg/config/webhook"
	eventhandlersvc "github.com/horizoncd/horizon/pkg/eventhandler"
	"github.com/horizoncd/horizon/pkg/eventhandler/wlgenerator"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	webhooksvc "github.com/horizoncd/horizon/pkg/webhook/service"
)

// Run runs the agent.
func Run(ctx context.Context, eventHandlerConfig eventhandlercfg.Config,
	webhookCfg webhookcfg.Config, mgrs *managerparam.Manager) {
	// start event handler service to generate webhook log by events
	eventHandlerService := eventhandlersvc.NewService(ctx, mgrs, eventHandlerConfig)
	if err := eventHandlerService.RegisterEventHandler("webhook",
		wlgenerator.NewWebhookLogGenerator(mgrs)); err != nil {
		log.Printf("failed to register event handler, error: %s", err.Error())
	}
	eventHandlerService.Start()

	// start webhook service with multi workers to consume webhook logs and send webhook events
	webhookService := webhooksvc.NewService(ctx, mgrs, webhookCfg)
	webhookService.Start()

	<-ctx.Done()
	// graceful exit
	webhookService.StopAndWait()
	eventHandlerService.StopAndWait()
}
