// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
