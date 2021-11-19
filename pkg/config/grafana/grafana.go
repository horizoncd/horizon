package grafana

type Mapper map[string]*Grafana

type Grafana struct {
	BasicDashboard      string `yaml:"basicDashboard"`
	ServerlessDashboard string `yaml:"serverlessDashboard"`
}
