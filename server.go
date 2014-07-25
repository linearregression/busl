package main

import (
	"bufio"
	"github.com/cyberdelia/pat"
	"io"
	"log"
	"net/http"
	"time"
)

func mkstream(w http.ResponseWriter, r *http.Request) {
	registrar := NewRedisRegistrar()
	uuid, err := NewUUID()
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
	uuid := UUID(r.URL.Query().Get(":uuid"))

	if !StringSliceUtil(r.TransferEncoding).Contains("chunked") {
		http.Error(w, "A chunked Transfer-Encoding header is required.", http.StatusBadRequest)
		return
	}

	msgBroker := NewRedisBroker(uuid)
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
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Transfer-Encoding", "chunked")

	uuid := UUID(r.URL.Query().Get(":uuid"))
	msgBroker := NewRedisBroker(uuid)
	ch, err := msgBroker.Subscribe()
	if err != nil {
		http.Error(w, "Channel is not registered.", http.StatusGone)
		f.Flush()
		return
	}

	defer msgBroker.UnsubscribeAll()

	ticker := time.NewTicker(time.Second * 20)

	for {
		select {
		case msg, msgOk := <- ch:
			if msgOk {
				w.Write(msg)
				f.Flush()
			} else {
				return
			}
		case _, tickOk := <- ticker.C:
			if tickOk {
				w.Header().Set("Hustle", "busl")
				f.Flush()
			} else {
				w.Write([]byte("Unable to keep connection alive."))
				f.Flush()
				return
			}
		}
	}

	ticker.Stop()
	f.Flush()
}

func Start() {
	p := pat.New()

	p.PostFunc("/streams", mkstream)
	p.PostFunc("/streams/:uuid", pub)
	p.GetFunc("/streams/:uuid", sub)

	http.Handle("/", p)

	if err := http.ListenAndServe(":"+*httpPort, nil); err != nil {
		panic(err)
	}
}
