package util

import (
	"log"
	"net/http"
	"strconv"
)

func NewResponseLogger(w http.ResponseWriter, requestID string) *responseLogger {
	return &responseLogger{ResponseWriter: w, status: http.StatusOK, requestID: requestID}
}

type responseLogger struct {
	http.ResponseWriter
	status    int
	requestID string
}

func (l *responseLogger) WriteHeader(s int) {
	l.ResponseWriter.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) CloseNotify() <-chan bool {
	return l.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (l *responseLogger) Flush() {
	l.ResponseWriter.(http.Flusher).Flush()
}

func (l *responseLogger) WriteLog() {
	maskedStatus := strconv.Itoa(l.status/100) + "xx"
	log.Printf("count#http.status.%s=1 status=%d request_id=%s", maskedStatus, l.status, l.requestID)
}
