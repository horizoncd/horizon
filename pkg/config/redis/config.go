package redis

// Redis
// ```yaml
// redisConfig:
//
//	protocol: tcp
//	location: 10.124.135.25
//	password: redis
//	db: 1
//
// ```
type Redis struct {
	Protocol string `yaml:"protocol"`
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       uint8  `yaml:"db"`
}
