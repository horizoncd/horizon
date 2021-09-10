package metrics

import (
	"fmt"
	"time"

	"g.hz.netease.com/horizon/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	_handlerLabel = "handler"
	_statusLabel  = "status"
)

var apiHistogram = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "horizon_api_request_histogram",
		Help:    "horizon api request histogram.",
		Buckets: prometheus.ExponentialBuckets(50, 2, 10),
	},
	[]string{_handlerLabel, _statusLabel},
)

// Middleware report metrics of handler
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// start timer
		start := time.Now()

		c.Next()

		// end timer
		latency := time.Now().Sub(start)

		statusCode := c.Writer.Status()
		handler := c.HandlerName()
		apiHistogram.With(prometheus.Labels{
			_handlerLabel: handler,
			_statusLabel: fmt.Sprintf("%v", statusCode),
		}).Observe(float64(latency.Milliseconds()))
	}, skippers...)
}

