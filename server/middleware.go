package server

import (
	"net/http"
)

func enforceHTTPS(ƒ http.HandlerFunc) http.HandlerFunc {
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
