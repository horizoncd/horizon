package grafana

type Mapper map[string]*Grafana

type Grafana struct {
	BasicDashboard      string `yaml:"basicDashboard"`
	ServerlessDashboard string `yaml:"serverlessDashboard"`
	MemcachedDashboard  string `yaml:"memcachedDashboard"`
	QuerySeries         string `yaml:"querySeries"`
}

type SLO struct {
	OverviewDashboard string                   `yaml:"overviewDashboard"`
	APIDashboard      string                   `yaml:"apiDashboard"`
	PipelineDashboard string                   `yaml:"pipelineDashboard"`
	APIReadRT         int                      `yaml:"readRT"`
	APIWriteRT        int                      `yaml:"writeRT"`
	GitRT             int                      `yaml:"gitRT"`
	ImageRT           int                      `yaml:"imageRT"`
	DeployRT          int                      `yaml:"deployRT"`
	Availability      map[string]*Availability `yaml:"availability"` // key is time range: 1h、1d、30d
}

type Availability struct {
	APIAvailability    float32 `yaml:"api"`
	GitAvailability    float32 `yaml:"git"`
	ImageAvailability  float32 `yaml:"image"`
	DeployAvailability float32 `yaml:"deploy"`
}
