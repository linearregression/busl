package server

import (
	"errors"
	"net/http"

	"github.com/heroku/busl/assets"
	"github.com/heroku/busl/broker"
	"github.com/heroku/busl/storage"
	"github.com/heroku/busl/util"
)

var errNoContent = errors.New("No Content")

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == broker.ErrNotRegistered || err == storage.ErrNoStorage {
		message := "Channel is not registered."
		if r.Header.Get("Accept") == "text/ascii; version=feral" {
			message = assets.HttpCatGone
		}

		http.Error(w, message, http.StatusNotFound)

	} else if err == errNoContent {
		// As indicated in the w3 spec[1] an SSE stream
		// that's already done should return a `204 No Content`
		// [1]: http://www.w3.org/TR/2012/WD-eventsource-20120426/
		w.WriteHeader(http.StatusNoContent)

	} else if err != nil {
		util.CountWithData("server.handleError", 1, "error=%s", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
