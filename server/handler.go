package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Handler is an http.Handler compatible interface which serves the busl api
type Handler struct {
	*Config
	Middleware
	*mux.Router
}

// NewHandler creates and configures a new handler instance
func NewHandler(config *Config) *Handler {
	h := &Handler{config, Middleware{config}, mux.NewRouter()}
	h.routes()
	return h
}

// ServeHTTP responds to HTTP requests with the right endpoint and middlewares
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Middleware.GetHandler(h.Router.ServeHTTP)(w, r)
}

func (h *Handler) routes() {
	e := Endpoint{h.Config}
	m := h.Middleware

	h.Router.HandleFunc("/health", m.WithoutAuth(e.Health))

	// Legacy endpoints for creating the uuid `key` for you.
	h.Router.HandleFunc("/streams", m.WithAuth(e.MakeUUID))

	// New `key` design for allowing any kind of id to be decided
	// by the caller (in this case, it mirrors what we have in S3).
	h.Router.HandleFunc("/streams/{key:.+}", m.WithoutAuth(e.Subscriber)).Methods("GET")
	h.Router.HandleFunc("/streams/{key:.+}", m.WithoutAuth(e.Publisher)).Methods("POST")
	h.Router.HandleFunc("/streams/{key:.+}", m.WithAuth(e.CreateStream)).Methods("PUT")
}
