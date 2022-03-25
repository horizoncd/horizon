package grafana

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
