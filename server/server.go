package server

import (
	"bufio"
	"github.com/cyberdelia/pat"
	"github.com/naaman/busl/assets"
	"github.com/naaman/busl/broker"
	"github.com/naaman/busl/util"
	"io"
	"log"
	"net/http"
	"time"
)

func mkstream(w http.ResponseWriter, _ *http.Request) {
	registrar := broker.NewRedisRegistrar()
	uuid, err := util.NewUUID()
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, "Unable to create stream. Please try again.", http.StatusServiceUnavailable)
		return
	}

	if err := registrar.Register(uuid); err != nil {
		log.Printf("%v", err)
		http.Error(w, "Unable to create stream. Please try again.", http.StatusServiceUnavailable)
		return
	}

	io.WriteString(w, string(uuid))
}

func pub(w http.ResponseWriter, r *http.Request) {
	uuid := util.UUID(r.URL.Query().Get(":uuid"))

	if !util.StringSliceUtil(r.TransferEncoding).Contains("chunked") {
		http.Error(w, "A chunked Transfer-Encoding header is required.", http.StatusBadRequest)
		return
	}

	msgBroker := broker.NewRedisBroker(uuid)
	defer msgBroker.UnsubscribeAll()

	scanner := bufio.NewScanner(r.Body)
	defer r.Body.Close()
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		msgBroker.Publish([]byte(scanner.Text() + "\n"))
	}
}

func sub(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	closeNotifier := w.(http.CloseNotifier).CloseNotify()

	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	uuid := util.UUID(r.URL.Query().Get(":uuid"))

	msgBroker := broker.NewRedisBroker(uuid)
	ch, err := msgBroker.Subscribe()
	defer msgBroker.Unsubscribe(ch)

	if err != nil {
		message := "Channel is not registered."
		if r.Header.Get("Accept") == "text/ascii; version=feral" {
			message = assets.HttpCatGone
		}

		http.Error(w, message, http.StatusGone)
		f.Flush()
		return
	}

	timer := time.NewTimer(*util.HeartbeatDuration)
	defer timer.Stop()

	for {
		select {
		case msg, msgOk := <-ch:
			if msgOk {
				timer.Reset(*util.HeartbeatDuration)
				w.Write(msg)
				f.Flush()
			} else {
				timer.Stop()
				return
			}
		case t, tOk := <-timer.C:
			if tOk {
				util.Count("server.sub.keepAlive")
				w.Write(util.GetNullByte())
				f.Flush()
				timer.Reset(*util.HeartbeatDuration)
			} else {
				util.CountWithData("server.sub.keepAlive.failed", 1, "timer=%v timerChannel=%v", timer, t)
				timer.Stop()
				w.Write([]byte("Unable to keep connection alive."))
				f.Flush()
				return
			}
		case cn, cnOk := <-closeNotifier:
			if cn && cnOk {
				util.Count("server.sub.clientClosed")
				timer.Stop()
				return
			}
		}
	}

	f.Flush()
}

func Start() {
	p := pat.New()

	p.PostFunc("/streams", mkstream)
	p.PostFunc("/streams/:uuid", pub)
	p.GetFunc("/streams/:uuid", sub)

	http.Handle("/", p)

	if err := http.ListenAndServe(":"+*util.HttpPort, nil); err != nil {
		panic(err)
	}
}
