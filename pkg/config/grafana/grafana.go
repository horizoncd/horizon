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
	URL                  string               `yaml:"url"`
	Namespace            string               `yaml:"namespace"`
	SyncDatasourceConfig SyncDatasourceConfig `yaml:"syncDatasourceConfig"`
	Dashboards           Dashboards           `yaml:"dashboards"`
}

type SyncDatasourceConfig struct {
	Period     time.Duration `yaml:"period"`
	LabelKey   string        `yaml:"labelKey"`
	LabelValue string        `yaml:"labelValue"`
}

type Dashboards struct {
	LabelKey   string `yaml:"labelKey"`
	LabelValue string `yaml:"labelValue"`
}
