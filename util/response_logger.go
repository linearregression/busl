package util

import (
	"log"
	"net/http"
	"strconv"
)

func NewResponseLogger(w http.ResponseWriter) *responseLogger {
	return &responseLogger{w: w, status: http.StatusOK}
}

type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size

	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) CloseNotify() <-chan bool {
	return l.w.(http.CloseNotifier).CloseNotify()
}

func (l *responseLogger) Flush() {
	l.w.(http.Flusher).Flush()
}

func (l *responseLogger) WriteLog(requestId string) {
	maskedStatus := strconv.Itoa(l.status/100) + "xx"
	log.Printf("count#http.status.%s=1 status=%d request_id=%s", maskedStatus, l.status, requestId)
}
