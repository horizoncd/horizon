// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package harbor

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var _harborDurationHistogram *prometheus.HistogramVec

const (
	_server     = "server"
	_method     = "method"
	_statuscode = "statuscode"
	_operation  = "operation"
)

func init() {
	_harborDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "harbor_request_duration_milliseconds",
		Help:    "Harbor request duration in milliseconds",
		Buckets: prometheus.ExponentialBuckets(50, 2, 10),
	}, []string{_server, _method, _statuscode, _operation})
}

func Observe(server, method, statuscode, operation string, duration time.Duration) {
	_harborDurationHistogram.With(prometheus.Labels{
		_server:     server,
		_method:     method,
		_statuscode: statuscode,
		_operation:  operation,
	}).Observe(float64(duration.Milliseconds()))
}
