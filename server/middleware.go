package server

import (
	"log"
	"net/http"

	"github.com/heroku/authenticater"
)

// Middleware holds the api middlewares used on top of the endpoints
type Middleware struct {
	*Config
}

// WithAuth returns the middlewares required for endpoints with authentication
func (m *Middleware) WithAuth(f http.HandlerFunc) http.HandlerFunc {
	return m.auth(m.WithoutAuth(f))
}

// WithoutAuth returns the middlewares required for endpoints without authentication
func (m *Middleware) WithoutAuth(f http.HandlerFunc) http.HandlerFunc {
	return m.addDefaultHeaders(f)
}

// GetHandler returns an http.Handler usable to serve default endpoints
func (m *Middleware) GetHandler(fn http.HandlerFunc) http.HandlerFunc {
	return m.logRequest(m.enforceHTTPS(fn))
}

func (m *Middleware) enforceHTTPS(fn http.HandlerFunc) http.HandlerFunc {
	if !m.EnforceHTTPS {
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

func (m *Middleware) auth(fn http.HandlerFunc) http.HandlerFunc {
	if m.Credentials == "" {
		return fn
	}

	auth, err := authenticater.NewBasicAuthFromString(m.Credentials)
	if err != nil {
		log.Fatalf("server.middleware error=%v", err)
		return nil
	}
	return authenticater.WrapAuth(auth, fn)
}

func (*Middleware) logRequest(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := newResponseLogger(r, w)
		fn(logger, r)
		logger.WriteLog()
	}
}

func (*Middleware) addDefaultHeaders(fn http.HandlerFunc) http.HandlerFunc {
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
