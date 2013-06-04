package eventsource

import (
	"net/http"
)

// Handler is an adapter for ordinary functions to act as an HTTP handler for
// event sources.
type Handler func(encoder *Encoder, stop <-chan bool)

// ServeHTTP calls h with an Encoder and a close notification channel. It
// performs Content-Type negotiation.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")

	if r.Header.Get("Accept") != "text/event-stream" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.WriteHeader(http.StatusOK)

	var stop <-chan bool

	if notifier, ok := w.(http.CloseNotifier); ok {
		stop = notifier.CloseNotify()
	}

	h(NewEncoder(w), stop)
}
