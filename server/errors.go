package server

import (
	"net/http"

	"github.com/heroku/busl/assets"
	"github.com/heroku/busl/broker"
)

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == broker.ErrNotRegistered {
		message := "Channel is not registered."
		if r.Header.Get("Accept") == "text/ascii; version=feral" {
			message = assets.HttpCatGone
		}

		http.Error(w, message, http.StatusGone)

	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
