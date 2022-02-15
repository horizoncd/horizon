package slo

type Dashboards struct {
	Overview string `json:"overview" yaml:"basic"`
	API      string `json:"api" yaml:"serverless,api"`
	Pipeline string `json:"pipeline" yaml:"memcached,pipeline"`
}
