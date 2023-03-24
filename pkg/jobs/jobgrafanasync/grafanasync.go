package jobgrafanasync

import (
	"context"

	"github.com/horizoncd/horizon/core/config"
	"github.com/horizoncd/horizon/pkg/grafana"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"k8s.io/client-go/kubernetes"
)

func Run(ctx context.Context, coreConfig *config.Config,
	manager *managerparam.Manager, client kubernetes.Interface) {
	grafanaService := grafana.NewService(coreConfig.GrafanaConfig, manager, client)
	grafanaService.SyncDatasource(ctx)
}
