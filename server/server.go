package server

import (
	"log"
	"time"

	"github.com/braintree/manners"
)

// Config holds all the server options
type Config struct {
	EnforceHTTPS      bool
	Credentials       string
	HeartbeatDuration time.Duration
	StorageBaseURL    string
}

// Server is a launchable api listener
type Server struct {
	*manners.GracefulServer
	*Config
}

// NewServer creates a new server instance
func NewServer(config *Config) *Server {
	return &Server{
		GracefulServer: manners.NewServer(),
		Config:         config,
	}
}

// Start starts the server instance
func (s *Server) Start(port string, shutdown <-chan struct{}) {
	log.Printf("http.start.port=%s\n", port)
	s.Handler = NewHandler(s.Config)
	go s.listenForShutdown(shutdown)

	s.Addr = ":" + port
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("server.server error=%v", err)
	}
}

func (s *Server) listenForShutdown(shutdown <-chan struct{}) {
	log.Println("http.graceful.await")
	<-shutdown
	log.Println("http.graceful.shutdown")
	s.Close()
}
