package server

import (
	"net/http"

	"github.com/heroku/busl/util"
)

func enforceHTTPS(ƒ http.HandlerFunc) http.HandlerFunc {
	if !*util.EnforceHTTPS {
		return ƒ
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "https" {
			url := r.URL
			url.Host = r.Host
			url.Scheme = "https"

			http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
			return
		}

		ƒ(w, r)
	}
}

func logRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := util.NewResponseLogger(w)
		fn(logger, r)
		logger.WriteLog()
	}
}
