package logging

import (
	"fmt"
	"log"
)

// Count parses a string into a count for logging to librato
func Count(metric string) { CountMany(metric, 1) }

// CountMany parses a string and number into a count for logging to librato
func CountMany(metric string, count int64) { CountWithData(metric, count, "") }

// CountWithData parses metrics for logging to librato
func CountWithData(metric string, count int64, extraData string, v ...interface{}) {
	if extraData == "" {
		log.Printf("count#%s=%d", metric, count)
	} else {
		log.Printf("count#%s=%d %s", metric, count, fmt.Sprintf(extraData, v...))
	}
}
