package db

type Config struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password,omitempty"`
	Database          string `yaml:"database"`
	PrometheusEnabled bool   `yaml:"prometheusEnabled"`
}
