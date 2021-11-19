package time

import "time"

const defaultTimeFormat = "2006-01-02 15:04:05"

func Now(timeFormat *string) string {
	var format = defaultTimeFormat
	if timeFormat != nil {
		format = *timeFormat
	}
	return time.Now().Format(format)
}
