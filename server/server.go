package server

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/braintree/manners"
	"github.com/cyberdelia/pat"
	"github.com/heroku/busl/assets"
	"github.com/heroku/busl/broker"
	"github.com/heroku/busl/sse"
	"github.com/heroku/busl/util"
	"github.com/heroku/rollbar"
)

var gracefulServer *manners.GracefulServer

func init() {
	gracefulServer = manners.NewServer()
	gracefulServer.InnerServer.ReadTimeout = *util.HttpReadTimeout
	gracefulServer.InnerServer.WriteTimeout = *util.HttpWriteTimeout
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

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func pub(w http.ResponseWriter, r *http.Request) {
	uuid := util.UUID(r.URL.Query().Get(":uuid"))

	if !util.StringSliceUtil(r.TransferEncoding).Contains("chunked") {
		http.Error(w, "A chunked Transfer-Encoding header is required.", http.StatusBadRequest)
		return
	}

	writer, err := broker.NewWriter(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	if err != nil {
		log.Printf("%#v", err)
		http.Error(w, "Unhandled error, please try again.", http.StatusInternalServerError)
		rollbar.Error(rollbar.ERR, fmt.Errorf("unhandled error: %#v", err))
	}
}

func sub(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	uuid := util.UUID(r.URL.Query().Get(":uuid"))

	rd, err := broker.NewReader(uuid)

	if err != nil {
		message := "Channel is not registered."
		if r.Header.Get("Accept") == "text/ascii; version=feral" {
			message = assets.HttpCatGone
		}

		http.Error(w, message, http.StatusGone)
		f.Flush()
		return
	}

	defer rd.Close()

	// Get the offset from Last-Event-ID: or Range:
	offset := offset(r)
	if offset > 0 {
		rd.(io.Seeker).Seek(int64(offset), 0)
	}

	// For default requests, we use a null byte for sending
	// the keepalive ack.
	ack := util.GetNullByte()

	if r.Header.Get("Accept") == "text/event-stream" {

		// As indicated in the w3 spec[1] an SSE stream
		// that's already done should return a 204
		// [1]: http://www.w3.org/TR/2012/WD-eventsource-20120426/
		if broker.ReaderDone(rd) {
			w.WriteHeader(http.StatusNoContent)
			f.Flush()
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		encoder := sse.NewEncoder(rd)
		encoder.(io.Seeker).Seek(int64(offset), 0)

		rd = ioutil.NopCloser(encoder)

		// For SSE, we change the ack to a :keepalive
		ack = []byte(":keepalive\n")
	}

	done := w.(http.CloseNotifier).CloseNotify()
	reader := NewKeepAliveReader(rd, ack, *util.HeartbeatDuration, done)
	io.Copy(NewWriteFlusher(w), reader)
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

func app() http.Handler {
	p := pat.New()

	p.GetFunc("/health", addDefaultHeaders(health))
	p.PostFunc("/streams", addDefaultHeaders(mkstream))
	p.PostFunc("/streams/:uuid", addDefaultHeaders(pub))
	p.GetFunc("/streams/:uuid", addDefaultHeaders(sub))

	return logRequest(enforceHTTPS(p.ServeHTTP))
}

func Start(port string, shutdown <-chan struct{}) {
	log.Printf("http.start.port=%s\n", port)
	http.Handle("/", app())
	go listenForShutdown(shutdown)

	if err := gracefulServer.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func listenForShutdown(shutdown <-chan struct{}) {
	log.Println("http.graceful.await")
	<-shutdown
	log.Println("http.graceful.shutdown")
	gracefulServer.InnerServer.SetKeepAlivesEnabled(false) // TODO: Remove after merge of https://github.com/braintree/manners/pull/22
	gracefulServer.Shutdown <- true
}
