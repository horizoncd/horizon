package job

type Config struct {
	// LockName is the name of the lock resource
	LockName string `yaml:"lockName"`

	// LockNS is the namespace of the lock resource
	LockNS string `yaml:"lockNS"`

	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership
	LeaseDuration int `yaml:"leaseDuration"`

	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up
	RenewDeadline int `yaml:"renewDeadline"`

	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions
	RetryPeriod int `yaml:"retryPeriod"`
}
