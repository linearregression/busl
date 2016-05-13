package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/braintree/manners"
	"github.com/gorilla/mux"
	"github.com/heroku/busl/broker"
	"github.com/heroku/busl/util"
	"github.com/heroku/rollbar"
)

var gracefulServer *manners.GracefulServer

func init() {
	gracefulServer = manners.NewServer()
	gracefulServer.ReadTimeout = *util.HTTPReadTimeout
	gracefulServer.WriteTimeout = *util.HTTPWriteTimeout
}

func mkstream(w http.ResponseWriter, _ *http.Request) {
	registrar := broker.NewRedisRegistrar()
	uuid, err := util.NewUUID()
	if err != nil {
		http.Error(w, "Unable to create stream. Please try again.", http.StatusServiceUnavailable)
		rollbar.Error(rollbar.ERR, fmt.Errorf("unable to create new uuid for stream: %#v", err))
		util.CountWithData("mkstream.create.fail", 1, "error=%s", err)
		return
	}

	if err := registrar.Register(uuid); err != nil {
		http.Error(w, "Unable to create stream. Please try again.", http.StatusServiceUnavailable)
		rollbar.Error(rollbar.ERR, fmt.Errorf("unable to register stream: %#v", err))
		util.CountWithData("mkstream.create.fail", 1, "error=%s", err)
		return
	}

	util.Count("mkstream.create.success")
	io.WriteString(w, string(uuid))
}

func put(w http.ResponseWriter, r *http.Request) {
	registrar := broker.NewRedisRegistrar()

	if err := registrar.Register(key(r)); err != nil {
		http.Error(w, "Unable to create stream. Please try again.", http.StatusServiceUnavailable)
		rollbar.Error(rollbar.ERR, fmt.Errorf("unable to register stream: %#v", err))
		util.CountWithData("put.create.fail", 1, "error=%s", err)
		return
	}
	util.Count("put.create.success")
	w.WriteHeader(http.StatusCreated)
}

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func pub(w http.ResponseWriter, r *http.Request) {
	if !util.StringInSlice(r.TransferEncoding, "chunked") {
		http.Error(w, "A chunked Transfer-Encoding header is required.", http.StatusBadRequest)
		return
	}

	writer, err := broker.NewWriter(key(r))
	if err != nil {
		handleError(w, r, err)
		return
	}
	defer writer.Close()

	body := bufio.NewReader(r.Body)
	defer r.Body.Close()

	_, err = io.Copy(writer, body)

	if err == io.ErrUnexpectedEOF {
		util.CountWithData("server.pub.read.eoferror", 1, "msg=\"%v\"", err.Error())
		return
	}

	netErr, ok := err.(net.Error)
	if ok && netErr.Timeout() {
		util.CountWithData("server.pub.read.timeout", 1, "msg=\"%v\"", err.Error())
		handleError(w, r, netErr)
		return
	}

	if err != nil {
		log.Printf("%#v", err)
		http.Error(w, "Unhandled error, please try again.", http.StatusInternalServerError)
		rollbar.Error(rollbar.ERR, fmt.Errorf("unhandled error: %#v", err))
		return
	}

	// Asynchronously upload the output to our defined storage backend.
	go storeOutput(key(r), requestURI(r))
}

func sub(w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	rd, err := newReader(w, r)
	if rd != nil {
		defer rd.Close()
	}
	if err != nil {
		handleError(w, r, err)
		return
	}
	_, err = io.Copy(newWriteFlusher(w), rd)

	netErr, ok := err.(net.Error)
	if ok && netErr.Timeout() {
		util.CountWithData("server.sub.read.timeout", 1, "msg=\"%v\"", err.Error())
		return
	}

	if err != nil {
		rollbar.Error(rollbar.ERR, fmt.Errorf("unhandled error: %#v", err))
	}
}

func app() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/health", addDefaultHeaders(health))

	// Legacy endpoint for creating the uuid `key` for you.
	r.HandleFunc("/streams", auth(addDefaultHeaders(mkstream)))

	// New `key` design for allowing any kind of id to be decided
	// by the caller (in this case, it mirrors what we have in S3).
	r.HandleFunc("/streams/{key:.+}", addDefaultHeaders(sub)).Methods("GET")
	r.HandleFunc("/streams/{key:.+}", addDefaultHeaders(pub)).Methods("POST")
	r.HandleFunc("/streams/{key:.+}", auth(addDefaultHeaders(put))).Methods("PUT")

	return logRequest(enforceHTTPS(r.ServeHTTP))
}

// Start starts the server instance
func Start(port string, shutdown <-chan struct{}) {
	log.Printf("http.start.port=%s\n", port)
	gracefulServer.Handler = app()
	go listenForShutdown(shutdown)

	gracefulServer.Addr = ":" + port
	if err := gracefulServer.ListenAndServe(); err != nil {
		log.Fatalf("server.server error=%v", err)
	}
}

func listenForShutdown(shutdown <-chan struct{}) {
	log.Println("http.graceful.await")
	<-shutdown
	log.Println("http.graceful.shutdown")
	gracefulServer.Close()
}
