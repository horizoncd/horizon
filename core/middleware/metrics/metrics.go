package metrics

import (
	"fmt"
	"regexp"
	"time"

	"g.hz.netease.com/horizon/pkg/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	_handlerLabel = "handler"
	_statusLabel  = "status"
	_verbLabel    = "verb"
)

var apiHistogram = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "horizon_api_request_duration_milliseconds",
		Help:    "horizon api request duration milliseconds histogram.",
		Buckets: prometheus.ExponentialBuckets(50, 2, 10),
	},
	[]string{_handlerLabel, _statusLabel, _verbLabel},
)

// Middleware report metrics of handler
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// start timer
		start := time.Now()

		c.Next()

		// end timer
		latency := time.Since(start)

		var handler string
		if handler = func() string {
			handlerName := c.HandlerName()
			// 忽略匿名的handler
			if regexp.MustCompile(`.*func\d*$`).MatchString(handlerName) {
				return ""
			}
			return handlerName
		}(); handler == "" {
			return
		}

		statusCode := c.Writer.Status()
		method := c.Request.Method

		apiHistogram.With(prometheus.Labels{
			_handlerLabel: handler,
			_statusLabel:  fmt.Sprintf("%v", statusCode),
			_verbLabel:    method,
		}).Observe(float64(latency.Milliseconds()))
	}, skippers...)
}
