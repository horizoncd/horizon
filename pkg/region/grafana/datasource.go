package grafana

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/util/kube"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
func CreatePrometheusDatasourceConfigMap(ctx context.Context, name string,
	labels map[string]string, datasource *DataSource) error {
	dsBytes, err := yaml.Marshal(datasource)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	_, client, err := kube.BuildClientFromContent("")
	if err != nil {
		return err
	}

	_, err = client.Basic.CoreV1().ConfigMaps("").Create(ctx, &v1.ConfigMap{
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
func UpdatePrometheusDatasourceConfigMap(ctx context.Context, name string,
	labels map[string]string, datasource *DataSource) error {
	dsBytes, err := yaml.Marshal(datasource)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	_, client, err := kube.BuildClientFromContent("")
	if err != nil {
		return err
	}

	_, err = client.Basic.CoreV1().ConfigMaps("").Update(ctx, &v1.ConfigMap{
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
func DeletePrometheusDatasourceConfigMap(ctx context.Context, name string) error {
	_, client, err := kube.BuildClientFromContent("")
	if err != nil {
		return err
	}

	err = client.Basic.CoreV1().ConfigMaps("").Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return perror.Wrap(herrors.ErrKubeDynamicCliResponseNotOK, err.Error())
	}

	return nil
}
