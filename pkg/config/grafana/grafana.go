package grafana

type Mapper map[string]*Grafana

type Grafana struct {
	BasicDashboard      string `yaml:"basicDashboard"`
	ServerlessDashboard string `yaml:"serverlessDashboard"`
	MemcachedDashboard  string `yaml:"memcachedDashboard"`
	QuerySeries         string `yaml:"querySeries"`
}

type SLO struct {
	APIDashboard      string                   `yaml:"apiDashboard"`
	PipelineDashboard string                   `yaml:"pipelineDashboard"`
	Availability      map[string]*Availability `yaml:"availability"`
}

type Availability struct {
	APIAvailability    float32 `yaml:"apiAvailability"`
	ReadRT             int     `yaml:"readRT"`
	WriteRT            int     `yaml:"writeRT"`
	GitAvailability    float32 `yaml:"gitAvailability"`
	GitRT              int     `yaml:"gitRT"`
	ImageAvailability  float32 `yaml:"imageAvailability"`
	ImageRT            int     `yaml:"imageRT"`
	DeployAvailability float32 `yaml:"deployAvailability"`
	DeployRT           int     `yaml:"deployRT"`
}
