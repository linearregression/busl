package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/satori/go.uuid"
)

func newResponseLogger(r *http.Request, w http.ResponseWriter) *responseLogger {
	return &responseLogger{ResponseWriter: w, request: r, status: http.StatusOK}
}

type responseLogger struct {
	http.ResponseWriter
	request *http.Request
	status  int
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

func (l *responseLogger) requestID() (id string) {
	if id = l.request.Header.Get("Request-Id"); id == "" {
		id = l.request.Header.Get("X-Request-Id")
	}
	if id == "" {
		id = uuid.NewV4().String()
	}

	return id
}

func (l *responseLogger) WriteLog() {
	maskedStatus := strconv.Itoa(l.status/100) + "xx"
	log.Printf("count#http.status.%s=1 request_id=%s", maskedStatus, l.requestID())
	log.Printf("method=%s path=\"%s\" host=\"%s\" fwd=\"%s\" status=%d user_agent=\"%s\" request_id=%s",
		l.request.Method, l.request.URL.Path, l.request.Host, l.request.Header.Get("X-Forwarded-For"), l.status, l.request.UserAgent(), l.requestID())
}
