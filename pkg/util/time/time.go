package time

import (
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultTimeFormat = "2006-01-02 15:04:05"

var cstSh, _ = time.LoadLocation("Asia/Shanghai")

func Now(timeFormat *string) string {
	var format = defaultTimeFormat
	if timeFormat != nil {
		format = *timeFormat
	}
	return time.Now().Format(format)
}

func K8sTimeToStrByNowTimezone(t v1.Time) string {
	return t.In(cstSh).Format(defaultTimeFormat)
}
