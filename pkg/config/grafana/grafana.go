package grafana

import "time"

type Mapper map[string]*Grafana

type Grafana struct {
	BasicDashboard      string `yaml:"basicDashboard"`
	ContainerDashboard  string `yaml:"containerDashboard"`
	ServerlessDashboard string `yaml:"serverlessDashboard"`
	MemcachedDashboard  string `yaml:"memcachedDashboard"`
	QuerySeries         string `yaml:"querySeries"`
}

type SLO struct {
	OverviewDashboard string `yaml:"overviewDashboard"`
	HistoryDashboard  string `yaml:"historyDashboard"`
}

type Config struct {
	GrafanaURL                   string        `yaml:"grafanaUrl"`
	DatasourceConfigMapNamespace string        `yaml:"datasourceConfigMapNamespace"`
	SyncDatasourcePeriod         time.Duration `yaml:"syncDatasourcePeriod"`
	SyncLockTTL                  time.Duration `yaml:"syncLockTTL"`
	Datasources                  Datasources   `yaml:"datasources"`
}

type Datasources struct {
	Label      string `yaml:"label"`
	LabelValue string `yaml:"labelValue"`
}
