package server

import (
	"io"
	"net/http"
)

// Use writeFlusher when you want io.Copy to keep
// flushing your http chunked response as it reads
// data.
type writeFlusher struct {
	w io.Writer
}

func NewWriteFlusher(w io.Writer) *writeFlusher {
	return &writeFlusher{w: w}
}

func (wf *writeFlusher) Write(p []byte) (int, error) {
	n, err := wf.w.Write(p)
	wf.w.(http.Flusher).Flush()
	return n, err
}
