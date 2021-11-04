package registry

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var _harborDurationHistogram *prometheus.HistogramVec

const (
	_server     = "server"
	_method     = "method"
	_uri        = "uri"
	_statuscode = "statuscode"
	_operation  = "operation"
)

func init() {
	_harborDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "harbor_request_duration_milliseconds",
		Help:    "Harbor request duration in milliseconds",
		Buckets: prometheus.ExponentialBuckets(50, 2, 10),
	}, []string{_server, _method, _uri, _statuscode, _operation})
}

func observe(server, method, uri, statuscode, operation string, duration time.Duration) {
	_harborDurationHistogram.With(prometheus.Labels{
		_server:     server,
		_method:     method,
		_uri:        uri,
		_statuscode: statuscode,
		_operation:  operation,
	}).Observe(float64(duration.Milliseconds()))
}
