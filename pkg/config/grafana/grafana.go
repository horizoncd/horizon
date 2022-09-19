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
	SyncDatasourceConfig SyncDatasourceConfig `yaml:"syncDatasourceConfig"`
}

type SyncDatasourceConfig struct {
	Namespace  string        `yaml:"namespace"`
	Period     time.Duration `yaml:"period"`
	LabelKey   string        `yaml:"labelKey"`
	LabelValue string        `yaml:"labelValue"`
}
