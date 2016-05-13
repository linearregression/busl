package server

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/heroku/authenticater"
	"github.com/heroku/busl/broker"
	"github.com/heroku/busl/encoders"
	"github.com/heroku/busl/logging"
	"github.com/heroku/busl/storage"
)

func (s *Server) enforceHTTPS(fn http.HandlerFunc) http.HandlerFunc {
	if !s.EnforceHTTPS {
		return fn
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "https" {
			url := r.URL
			url.Host = r.Host
			url.Scheme = "https"

			http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
			return
		}

		fn(w, r)
	}
}

func (s *Server) auth(fn http.HandlerFunc) http.HandlerFunc {
	if s.Credentials == "" {
		return fn
	}

	auth, err := authenticater.NewBasicAuthFromString(s.Credentials)
	if err != nil {
		log.Fatalf("server.middleware error=%v", err)
		return nil
	}
	return authenticater.WrapAuth(auth, fn)
}

func logRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := newResponseLogger(r, w)
		fn(logger, r)
		logger.WriteLog()
	}
}

func (s *Server) addDefaultHeaders(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		fn(w, r)
	}
}

func offset(r *http.Request) int64 {
	var off string

	if off = r.Header.Get("last-event-id"); off == "" {
		if val := r.Header.Get("Range"); val != "" {
			tuple := strings.SplitN(val, "-", 2)
			off = tuple[0]
		}
	}

	n, _ := strconv.Atoi(off)
	return int64(n)
}

// Given URL:
//   http://build-output.heroku.com/streams/1/2/3?foo=bar
//
// Returns:
//   1/2/3?foo=bar
func requestURI(r *http.Request) string {
	res := key(r)

	if r.URL.RawQuery != "" {
		res += "?" + r.URL.RawQuery
	}

	return res
}

func key(r *http.Request) string {
	return mux.Vars(r)["key"]
}

// Returns a broker or blob reader.
func (s *Server) newStorageReader(w http.ResponseWriter, r *http.Request) (io.ReadCloser, error) {
	// Get the offset from Last-Event-ID: or Range:
	offset := offset(r)

	rd, err := broker.NewReader(key(r))

	// Not cached in the broker anymore, try the storage backend as a fallback.
	if err == broker.ErrNotRegistered {
		return storage.Get(requestURI(r), s.StorageBaseURL, offset)
	}

	if offset > 0 {
		if seeker, ok := rd.(io.Seeker); ok {
			seeker.Seek(offset, 0)
		}
	}
	return rd, err
}

func (s *Server) newReader(w http.ResponseWriter, r *http.Request) (io.ReadCloser, error) {
	rd, err := s.newStorageReader(w, r)
	if err != nil {
		if rd != nil {
			rd.Close()
		}
		return rd, err
	}

	// For default requests, we use a null byte for sending
	// the keepalive ack.
	ack := []byte{0}

	if broker.NoContent(rd, offset(r)) {
		return nil, errNoContent
	}

	if r.Header.Get("Accept") == "text/event-stream" {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		encoder := encoders.NewSSEEncoder(rd)
		encoder.(io.Seeker).Seek(offset(r), 0)

		rd = ioutil.NopCloser(encoder)

		// For SSE, we change the ack to a :keepalive
		ack = []byte(":keepalive\n")
	}

	done := w.(http.CloseNotifier).CloseNotify()
	return newKeepAliveReader(rd, ack, s.HeartbeatDuration, done), nil
}

func storeOutput(channel string, requestURI string, storageBase string) {
	if buf, err := broker.Get(channel); err == nil {
		if err := storage.Put(requestURI, storageBase, bytes.NewBuffer(buf)); err != nil {
			logging.CountWithData("server.storeOutput.put.error", 1, "err=%s", err.Error())
		}
	} else {
		logging.CountWithData("server.storeOutput.get.error", 1, "err=%s", err.Error())
	}
}
