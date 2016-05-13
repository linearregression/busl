package server

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/heroku/busl/broker"
	"github.com/heroku/busl/encoders"
	"github.com/heroku/busl/logging"
	"github.com/heroku/busl/storage"
)

func hasEncoding(encodings []string, check string) bool {
	for _, c := range encodings {
		if c == check {
			return true
		}
	}
	return false
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

func newReader(c *Config, w http.ResponseWriter, r *http.Request) (io.ReadCloser, error) {
	rd, err := newStorageReader(c.StorageBaseURL, w, r)
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
	return newKeepAliveReader(rd, ack, c.HeartbeatDuration, done), nil
}

func newStorageReader(storageBaseURL string, w http.ResponseWriter, r *http.Request) (io.ReadCloser, error) {
	// Get the offset from Last-Event-ID: or Range:
	offset := offset(r)

	rd, err := broker.NewReader(key(r))

	// Not cached in the broker anymore, try the storage backend as a fallback.
	if err == broker.ErrNotRegistered {
		return storage.Get(requestURI(r), storageBaseURL, offset)
	}

	if offset > 0 {
		if seeker, ok := rd.(io.Seeker); ok {
			seeker.Seek(offset, 0)
		}
	}
	return rd, err
}

func key(r *http.Request) string {
	return mux.Vars(r)["key"]
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
