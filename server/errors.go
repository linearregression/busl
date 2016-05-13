package server

import (
	"errors"
	"net/http"

	"github.com/heroku/busl/broker"
	"github.com/heroku/busl/logging"
	"github.com/heroku/busl/storage"
)

var errNoContent = errors.New("No Content")

const asciiGone = `░░░░░░░░░░██░░░░░░░░░░██░░░░░░░░
░░░░░░░░██░░██░░░░░░██░░██░░░░░░
░░░░░░░░██░░░░██████░░░░██░░░░░░
░░░░░░░░██░░░░░░░░░░░░░░██░░░░░░
░░░░██████░░██░░░░░░██░░██████░░
░░░░██░░░░░░██░░░░░░██░░░░░░██░░
░░░░░░██░░░░░░░░██░░░░░░░░██░░░░
░░░░░░░░██░░░░██░░██░░░░██░░░░░░
░░░░░░░░░░██████████████░░░░░░░░
░░░░░░░░░░██░░░░░░░░██░░░░░░░░░░
░░░░░░░░██░░░░░░░░████░░░GONE░░░
░░░░░░░░██░░░░░░░░██░░░░░░░░░░░░
██░░░░██░░░░░░██░░██░░░░░░░░░░░░
██░░░░██░░██░░██░░████░░░░░░░░░░
██░░██░░░░██░░██░░██░░██░░░░░░░░
░░░░██░░░░██░░██░░██░░██░░░░░░░░`

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch err {
	case broker.ErrNotRegistered, storage.ErrNoStorage, storage.ErrNotFound:
		message := "Channel is not registered."
		if r.Header.Get("Accept") == "text/ascii; version=feral" {
			message = asciiGone
		}

		http.Error(w, message, http.StatusNotFound)

	case storage.ErrRange:
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)

	case errNoContent:
		// As indicated in the w3 spec[1] an SSE stream
		// that's already done should return a `204 No Content`
		// [1]: http://www.w3.org/TR/2012/WD-eventsource-20120426/
		w.WriteHeader(http.StatusNoContent)

	default:
		logging.CountWithData("server.handleError", 1, "error=%s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}
