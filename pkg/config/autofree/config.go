package autofree

import "time"

type Config struct {
	SupportedEnvs []string      `yaml:"supportedEnvs"`
	Account       string        `yaml:"account"`
	JobInterval   time.Duration `yaml:"jobInterval"`
	BatchInterval time.Duration `yaml:"batchInterval"`
	BatchSize     int           `yaml:"batchSize"`
}
