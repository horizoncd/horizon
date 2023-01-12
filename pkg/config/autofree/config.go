package autofree

import "time"

type Config struct {
	SupportedEnvs []string      `yaml:"supportedEnvs"`
	AccountID     uint          `yaml:"accountID"`
	JobInterval   time.Duration `yaml:"jobInterval"`
	BatchInterval time.Duration `yaml:"batchInterval"`
	BatchSize     int           `yaml:"batchSize"`
}
