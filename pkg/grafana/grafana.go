package grafana

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/config/grafana"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/util/kube"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Service interface {
	CreatePrometheusDatasourceConfigMap(ctx context.Context, name string,
		labels map[string]string, datasource *DataSource) error
	UpdatePrometheusDatasourceConfigMap(ctx context.Context, name string,
		labels map[string]string, datasource *DataSource) error
	DeletePrometheusDatasourceConfigMap(ctx context.Context, name string) error
	GetPrometheusDatasourceConfigMap(ctx context.Context, name string) (*DataSource, error)
}

type service struct {
	config grafana.Config
	client kubernetes.Interface
}

func NewService(config grafana.Config, client kubernetes.Interface) Service {
	return &service{
		config: config,
		client: client,
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

const (
	DatasourceKey = "datasource.yaml"
)

// CreatePrometheusDatasourceConfigMap create a prometheus datasource for the grafana
func (s *service) CreatePrometheusDatasourceConfigMap(ctx context.Context, name string,
	labels map[string]string, datasource *DataSource) error {
	dsBytes, err := yaml.Marshal(datasource)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	_, err = s.client.CoreV1().ConfigMaps(s.config.DatasourceConfigMapNamespace).Create(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Data: map[string]string{
			DatasourceKey: string(dsBytes),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrKubeDynamicCliResponseNotOK, err.Error())
	}

	return nil
}

// UpdatePrometheusDatasourceConfigMap update a prometheus datasource for the grafana
func (s *service) UpdatePrometheusDatasourceConfigMap(ctx context.Context, name string,
	labels map[string]string, datasource *DataSource) error {
	dsBytes, err := yaml.Marshal(datasource)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	_, client, err := kube.BuildClientFromContent("")
	if err != nil {
		return err
	}

	_, err = client.Basic.CoreV1().ConfigMaps(s.config.DatasourceConfigMapNamespace).Update(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Data: map[string]string{
			DatasourceKey: string(dsBytes),
		},
	}, metav1.UpdateOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrKubeDynamicCliResponseNotOK, err.Error())
	}

	return nil
}

// DeletePrometheusDatasourceConfigMap delete a prometheus datasource for the grafana
func (s *service) DeletePrometheusDatasourceConfigMap(ctx context.Context, name string) error {
	err := s.client.CoreV1().ConfigMaps(s.config.DatasourceConfigMapNamespace).Delete(ctx, name,
		metav1.DeleteOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrKubeDynamicCliResponseNotOK, err.Error())
	}

	return nil
}

// GetPrometheusDatasourceConfigMap delete a prometheus datasource for the grafana
func (s *service) GetPrometheusDatasourceConfigMap(ctx context.Context, name string) (*DataSource, error) {
	configmap, err := s.client.CoreV1().ConfigMaps(s.config.DatasourceConfigMapNamespace).Get(ctx, name,
		metav1.GetOptions{})
	if err != nil {
		return nil, perror.Wrap(herrors.ErrKubeDynamicCliResponseNotOK, err.Error())
	}

	data, ok := configmap.Data[DatasourceKey]
	if ok {
		var ds *DataSource
		err = yaml.Unmarshal([]byte(data), ds)
		if err != nil {
			return nil, perror.WithMessage(herrors.ErrGrafanaDatasourceFormat, err.Error())
		}
		return ds, nil
	}
	return nil, herrors.ErrGrafanaDatasourceFormat
}
