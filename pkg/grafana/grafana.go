package grafana

import (
	"context"
	"fmt"
	"net/http"
	"time"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/config/grafana"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	DatasourceKey                 = "datasource.yaml"
	DatasourceConfigMapNamePrefix = "grafana-datasource"
	PrometheusDatasourceType      = "prometheus"
	SyncDatasourceLockKey         = "sync_datasource_lock"
)

type Service interface {
	SyncDatasource(ctx context.Context)
}

type service struct {
	config      grafana.Config
	kubeClient  kubernetes.Interface
	regionMgr   regionmanager.Manager
	redisClient *redis.Client
}

func NewService(config grafana.Config, manager *managerparam.Manager, client kubernetes.Interface,
	redisClient *redis.Client) Service {
	return &service{
		config:      config,
		kubeClient:  client,
		regionMgr:   manager.RegionMgr,
		redisClient: redisClient,
	}
}

type DataSource struct {
	OrgID   int64
	Version int

	Name            string
	Type            string
	Access          string
	URL             string
	User            string
	Database        string
	BasicAuth       bool
	BasicAuthUser   string
	WithCredentials bool
	IsDefault       bool
	Correlations    []map[string]interface{}
	JSONData        map[string]interface{}
	SecureJSONData  map[string]string
	Editable        bool
	UID             string
}

func (s *service) SyncDatasource(ctx context.Context) {
	log.Infof(ctx, "Starting syncing grafana datasource every %v", s.config.SyncDatasourcePeriod)
	defer log.Infof(ctx, "Stopping syncing grafana datasource")

	ticker := time.NewTicker(s.config.SyncDatasourcePeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug(ctx, "get done signal from context")
			return
		case <-ticker.C:
		}

		s.sync(ctx)
	}
}

func (s *service) sync(ctx context.Context) {
	resp := s.redisClient.SetNX(ctx, SyncDatasourceLockKey, 1, s.config.SyncLockTTL)
	lockSuccess, err := resp.Result()

	if err != nil {
		log.Errorf(ctx, "sync datasource error: %+v", err)
		return
	}

	if !lockSuccess {
		log.Debug(ctx, "get lock failed")
		return
	}

	log.Debug(ctx, "get lock succeed")
	regions, err := s.regionMgr.ListAll(ctx)
	if err != nil {
		log.Errorf(ctx, "sync datasource error: %+v", err)
		return
	}

	rMap := map[string]int{}
	// iterate regions, create or update configmap
	for _, region := range regions {
		configMapName := formatDatasourceConfigMapName(region.Name)
		datasourceURL := region.PrometheusURL
		ds := &DataSource{
			Name: region.Name,
			Type: PrometheusDatasourceType,
			URL:  datasourceURL,
		}
		err = s.getPrometheusDatasourceConfigMap(ctx, configMapName)
		if err != nil {
			// configmap not found, just create it
			if statusError, ok := err.(*k8serrors.StatusError); ok && statusError.ErrStatus.Code == http.StatusNotFound {
				err := s.createPrometheusDatasourceConfigMap(ctx, configMapName, ds)
				if err != nil {
					log.Errorf(ctx, "sync datasource error: %+v", err)
				}
			} else {
				log.Errorf(ctx, "sync datasource error: %+v", err)
			}
		} else {
			// found, just update it
			err := s.updatePrometheusDatasourceConfigMap(ctx, configMapName, ds)
			if err != nil {
				log.Errorf(ctx, "sync datasource error: %+v", err)
			}
		}

		rMap[configMapName] = 1
	}

	// iterate configmaps, delete unused ones
	configMaps, err := s.getPrometheusDatasourceConfigMaps(ctx)
	if err != nil {
		log.Errorf(ctx, "sync datasource error: %+v", err)
		return
	}

	for _, item := range configMaps.Items {
		if _, ok := rMap[item.Name]; !ok {
			// delete unused ones
			err = s.deletePrometheusDatasourceConfigMap(ctx, item.Name)
			if err != nil {
				log.Errorf(ctx, "sync datasource error: %+v", err)
			}
		}
	}
}

// createPrometheusDatasourceConfigMap create a prometheus datasource for the grafana
func (s *service) createPrometheusDatasourceConfigMap(ctx context.Context, name string, datasource *DataSource) error {
	dsBytes, err := yaml.Marshal(datasource)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	_, err = s.kubeClient.CoreV1().ConfigMaps(s.config.DatasourceConfigMapNamespace).Create(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				s.config.Datasources.Label: s.config.Datasources.LabelValue,
			},
		},
		Data: map[string]string{
			DatasourceKey: string(dsBytes),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrAPIServerResponseNotOK, err.Error())
	}

	return nil
}

// updatePrometheusDatasourceConfigMap update a prometheus datasource for the grafana
func (s *service) updatePrometheusDatasourceConfigMap(ctx context.Context, name string, datasource *DataSource) error {
	dsBytes, err := yaml.Marshal(datasource)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	_, err = s.kubeClient.CoreV1().ConfigMaps(s.config.DatasourceConfigMapNamespace).Update(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				s.config.Datasources.Label: s.config.Datasources.LabelValue,
			},
		},
		Data: map[string]string{
			DatasourceKey: string(dsBytes),
		},
	}, metav1.UpdateOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrAPIServerResponseNotOK, err.Error())
	}

	return nil
}

// deletePrometheusDatasourceConfigMap delete a prometheus datasource for the grafana
func (s *service) deletePrometheusDatasourceConfigMap(ctx context.Context, name string) error {
	err := s.kubeClient.CoreV1().ConfigMaps(s.config.DatasourceConfigMapNamespace).Delete(ctx, name,
		metav1.DeleteOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrAPIServerResponseNotOK, err.Error())
	}

	return nil
}

// GetPrometheusDatasourceConfigMap get a prometheus datasource for the grafana
func (s *service) getPrometheusDatasourceConfigMap(ctx context.Context, name string) error {
	_, err := s.kubeClient.CoreV1().ConfigMaps(s.config.DatasourceConfigMapNamespace).Get(ctx, name,
		metav1.GetOptions{})
	if err != nil {
		return err
	}

	return nil
}

// getPrometheusDatasourceConfigMaps get prometheus datasource for the grafana
func (s *service) getPrometheusDatasourceConfigMaps(ctx context.Context) (*v1.ConfigMapList, error) {
	datasources, err := s.kubeClient.CoreV1().ConfigMaps(s.config.DatasourceConfigMapNamespace).List(ctx,
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%v=%v", s.config.Datasources.Label, s.config.Datasources.LabelValue),
		})
	if err != nil {
		return nil, err
	}

	return datasources, nil
}

func formatDatasourceConfigMapName(name string) string {
	return fmt.Sprintf("%s-%s", DatasourceConfigMapNamePrefix, name)
}
