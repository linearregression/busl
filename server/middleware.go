package server

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/heroku/busl/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/heroku/busl/Godeps/_workspace/src/github.com/heroku/authenticater"
	"github.com/heroku/busl/util"
)

func enforceHTTPS(fn http.HandlerFunc) http.HandlerFunc {
	if !*util.EnforceHTTPS {
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

func auth(fn http.HandlerFunc) http.HandlerFunc {
	if *util.Creds == "" {
		return fn
	}

	if auth, err := authenticater.NewBasicAuthFromString(*util.Creds); err != nil {
		log.Fatalf("server.middleware error=%v", err)
		return nil
	} else {
		return authenticater.WrapAuth(auth, fn)
	}
}

func logRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := util.NewResponseLogger(w, requestId(r))
		fn(logger, r)
		logger.WriteLog()
	}
}

func addDefaultHeaders(fn http.HandlerFunc) http.HandlerFunc {
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

func requestId(r *http.Request) (id string) {
	if id = r.Header.Get("Request-Id"); id == "" {
		id = r.Header.Get("X-Request-Id")
	}

	if id == "" {
		// In the event of a rare case where uuid
		// generation fails, it's probably more
		// desirable to continue as is with an empty
		// request_id than to bubble the error up the
		// stack.
		uuid, _ := util.NewUUID()
		id = string(uuid)
	}

	return id
}

func offset(r *http.Request) (n int) {
	var off string

	if off = r.Header.Get("last-event-id"); off == "" {
		if val := r.Header.Get("Range"); val != "" {
			tuple := strings.SplitN(val, "-", 2)
			off = tuple[0]
		}
	}

	n, _ = strconv.Atoi(off)
	return n
}

func key(r *http.Request) string {
	return mux.Vars(r)["key"]
}
