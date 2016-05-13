package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/heroku/busl/broker"
	"github.com/heroku/busl/logging"
	"github.com/heroku/rollbar"
	"github.com/satori/go.uuid"
)

// Endpoint holds the api endpoints
type Endpoint struct {
	*Config
}

// MakeUUID is a legacy endpoint used to create a uuid key
func (e *Endpoint) MakeUUID(w http.ResponseWriter, _ *http.Request) {
	registrar := broker.NewRedisRegistrar()
	uuid := uuid.NewV4().String()

	if err := registrar.Register(uuid); err != nil {
		http.Error(w, "Unable to create stream. Please try again.", http.StatusServiceUnavailable)
		rollbar.Error(rollbar.ERR, fmt.Errorf("unable to register stream: %#v", err))
		logging.CountWithData("mkstream.create.fail", 1, "error=%s", err)
		return
	}

	logging.Count("mkstream.create.success")
	io.WriteString(w, string(uuid))
}

// CreateStream creates a new stream
func (e *Endpoint) CreateStream(w http.ResponseWriter, r *http.Request) {
	registrar := broker.NewRedisRegistrar()

	if err := registrar.Register(key(r)); err != nil {
		http.Error(w, "Unable to create stream. Please try again.", http.StatusServiceUnavailable)
		rollbar.Error(rollbar.ERR, fmt.Errorf("unable to register stream: %#v", err))
		logging.CountWithData("put.create.fail", 1, "error=%s", err)
		return
	}
	logging.Count("put.create.success")
	w.WriteHeader(http.StatusCreated)
}

// Health checks app's health
func (e *Endpoint) Health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

// Publisher allows sending data to a stream
func (e *Endpoint) Publisher(w http.ResponseWriter, r *http.Request) {
	if !hasEncoding(r.TransferEncoding, "chunked") {
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
		logging.CountWithData("server.pub.read.eoferror", 1, "msg=\"%v\"", err.Error())
		return
	}

	netErr, ok := err.(net.Error)
	if ok && netErr.Timeout() {
		logging.CountWithData("server.pub.read.timeout", 1, "msg=\"%v\"", err.Error())
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
	go storeOutput(key(r), requestURI(r), e.StorageBaseURL)
}

// Subscriber allows reading data from a stream
func (e *Endpoint) Subscriber(w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	rd, err := newReader(e.Config, w, r)
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
		logging.CountWithData("server.sub.read.timeout", 1, "msg=\"%v\"", err.Error())
		return
	}

	if err != nil {
		rollbar.Error(rollbar.ERR, fmt.Errorf("unhandled error: %#v", err))
	}
}
