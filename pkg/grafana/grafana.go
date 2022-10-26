package grafana

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/config/grafana"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	"g.hz.netease.com/horizon/pkg/util/log"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	_datasourceConfigMapName  = "grafana-datasource"
	_prometheusDatasourceType = "prometheus"
	_contentMD5AnnotationKey  = "content-md5"
	_datasourceDataKey        = "horizon-datasource.yaml"
	_datasourceAPIVersion     = 1
)

type Service interface {
	SyncDatasource(ctx context.Context)
	ListDashboards(ctx context.Context) ([]*Dashboard, error)
}

type service struct {
	config     grafana.Config
	kubeClient kubernetes.Interface
	regionMgr  regionmanager.Manager
}

func NewService(config grafana.Config, manager *managerparam.Manager, client kubernetes.Interface) Service {
	return &service{
		config:     config,
		kubeClient: client,
		regionMgr:  manager.RegionMgr,
	}
}

type Content struct {
	APIVersion  int          `yaml:"apiVersion"`
	Datasources []DataSource `yaml:"datasources"`
}

type DataSource struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // only use prometheus currently
	URL  string `yaml:"url"`
}

// Dashboard used to unmarshal grafana dashboard's content
type Dashboard struct {
	Title string   `json:"title"`
	UID   string   `json:"uid"`
	Tags  []string `json:"tags"`
}

func (s *service) SyncDatasource(ctx context.Context) {
	log.Infof(ctx, "Starting syncing grafana datasource every %v", s.config.SyncDatasourceConfig.Period)
	defer log.Infof(ctx, "Stopping syncing grafana datasource")

	ticker := time.NewTicker(s.config.SyncDatasourceConfig.Period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug(ctx, "Get done signal from context")
			return
		case <-ticker.C:
		}

		s.sync(ctx)
	}
}

func (s *service) sync(ctx context.Context) {
	log.Info(ctx, "Start to sync grafana datasource")

	logErr := func(err error) {
		log.Errorf(ctx, "Sync grafana datasource error: %+v", err)
	}

	regions, err := s.regionMgr.ListAll(ctx)
	if err != nil {
		logErr(err)
		return
	}

	configMapOps := s.kubeClient.CoreV1().ConfigMaps(s.config.Namespace)
	datasourceConfigMap, err := configMapOps.Get(ctx, _datasourceConfigMapName, metav1.GetOptions{})
	if err != nil {
		if statusError, ok := err.(*k8serrors.StatusError); !ok || statusError.ErrStatus.Code != http.StatusNotFound {
			logErr(err)
			return
		}
	}

	var datasources []DataSource
	for _, region := range regions {
		datasourceURL := region.PrometheusURL
		datasources = append(datasources, DataSource{
			Name: region.Name,
			Type: _prometheusDatasourceType,
			URL:  datasourceURL,
		})
	}

	content := Content{
		APIVersion:  _datasourceAPIVersion,
		Datasources: datasources,
	}
	dsBytes, err := yaml.Marshal(&content)
	if err != nil {
		logErr(err)
		return
	}
	h := md5.New()
	h.Write(dsBytes)
	curMD5Val := hex.EncodeToString(h.Sum(nil))
	curConfigmap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: _datasourceConfigMapName,
			Labels: map[string]string{
				s.config.SyncDatasourceConfig.LabelKey: s.config.SyncDatasourceConfig.LabelValue,
			},
			Annotations: map[string]string{
				_contentMD5AnnotationKey: curMD5Val,
			},
		},
		Data: map[string]string{
			_datasourceDataKey: string(dsBytes),
		},
	}

	// datasourceConfigMap is not nil though it is a pointer because the client-go's logic
	contentMD5, ok := datasourceConfigMap.ObjectMeta.Annotations[_contentMD5AnnotationKey]
	if !ok {
		// create configmap
		_, err := configMapOps.Create(ctx, curConfigmap, metav1.CreateOptions{})
		if err != nil {
			logErr(err)
		} else {
			log.Infof(ctx, "Create grafana datasource successfully, content: %s", string(dsBytes))
		}
		return
	}

	// update the configmap if md5 values are not equal.
	if contentMD5 != curMD5Val {
		_, err := configMapOps.Update(ctx, curConfigmap, metav1.UpdateOptions{})
		if err != nil {
			logErr(err)
		} else {
			log.Infof(ctx, "Update grafana datasource successfully, content: %s", string(dsBytes))
		}
		return
	}

	log.Debug(ctx, "Skip updating datasource because there are no changes")
}

func (s *service) ListDashboards(ctx context.Context) ([]*Dashboard, error) {
	configMapOps := s.kubeClient.CoreV1().ConfigMaps(s.config.Namespace)

	dashboardConfigMapList, err := configMapOps.List(ctx,
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%v=%v", s.config.Dashboards.LabelKey,
				s.config.Dashboards.LabelValue),
		})
	if err != nil {
		if statusError, ok := err.(*k8serrors.StatusError); !ok || statusError.ErrStatus.Code != http.StatusNotFound {
			return nil, perror.Wrap(herrors.ErrListGrafanaDashboard, err.Error())
		}
	}

	var dashboards []*Dashboard
	if dashboardConfigMapList != nil {
		for _, item := range dashboardConfigMapList.Items {
			for _, val := range item.Data {
				var dashboard Dashboard
				err = json.Unmarshal([]byte(val), &dashboard)
				if err != nil {
					return nil, perror.WithMessage(herrors.ErrListGrafanaDashboard, err.Error())
				}
				dashboards = append(dashboards, &dashboard)
			}
		}
	}

	return dashboards, nil
}
