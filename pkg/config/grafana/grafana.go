package grafana

import "time"

type SLO struct {
	OverviewDashboard string `yaml:"overviewDashboard"`
	HistoryDashboard  string `yaml:"historyDashboard"`
}

type Config struct {
	Host                 string               `yaml:"host"`
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
