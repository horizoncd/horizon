package webhook

type Config struct {
	// seconds for http client timeout
	ClientTimeout uint `yaml:"clientTimeout"`
	// seconds to wait when there is no log to proress
	IdleWaitInterval uint `yaml:"idleWaitInterval"`
	// seconds to wait after complete a worker reconciliation
	WorkerReconcileInterval uint `yaml:"workerReconcileInterval"`
	// bytes limit to truncate for response body
	ResponseBodyTruncateSize uint `yaml:"responseBodyTruncateSize"`
}
