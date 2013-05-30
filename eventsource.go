// Package eventsource provides the building blocks for consuming and building
// EventSource services.
package eventsource

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"
)

var (
	ErrClosed = errors.New("closed")
)

type Message struct {
	ID    []byte
	Event []byte
	Data  []byte
}

// An EventSource consumes server sent events over HTTP with automatic
// recovery.
type EventSource struct {
	retry       time.Duration
	request     *http.Request
	err         error
	r           io.ReadCloser
	dec         *Decoder
	lastEventId []byte
}

// Prepare an EventSource. The connection is automatically managed, using req
// to connect, and retrying from recoverable errors after waiting the provided
// retry duration.
func New(req *http.Request, retry time.Duration) *EventSource {
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	return &EventSource{
		retry:   retry,
		request: req,
	}
}

// Close the source. Any further calls to Read() will return ErrClosed.
func (es *EventSource) Close() {
	if es.r != nil {
		es.r.Close()
	}
	es.err = ErrClosed
}

// Connect to an event source, validate the response, and gracefully handle
// reconnects.
func (es *EventSource) connect() {
	for es.err == nil {
		if es.r != nil {
			es.r.Close()
			<-time.After(es.retry)
		}

		es.request.Header.Set("Last-Event-Id", string(es.lastEventId))

		resp, err := http.DefaultClient.Do(es.request)

		if err != nil {
			continue // reconnect
		}

		if resp.StatusCode >= 500 {
			// assumed to be temporary, try reconnecting
			resp.Body.Close()
		} else if resp.StatusCode == 204 {
			resp.Body.Close()
			es.err = ErrClosed
		} else if resp.StatusCode != 200 {
			resp.Body.Close()
			es.err = fmt.Errorf("endpoint returned unrecoverable status %q", resp.Status)
		} else {
			mediatype, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))

			if mediatype != "event/text-stream" {
				resp.Body.Close()
				es.err = fmt.Errorf("invalid content type %q", resp.Header.Get("Content-Type"))
			} else {
				es.r = resp.Body
				es.dec = NewDecoder(es.r)
				return
			}
		}
	}
}

// Read a message from EventSource. If an error is returned, the EventSource
// will not reconnect, and any further call to Read() will return the same
// error.
func (es *EventSource) Read() (Message, error) {
	if es.r == nil {
		es.connect()
	}

	for es.err == nil {
		id, event, data, err := es.dec.Read()

		if err != nil {
			es.connect()
			continue
		}

		if len(data) == 0 {
			continue
		}

		if len(event) == 0 {
			event = []byte("message")
		}

		if id != nil {
			es.lastEventId = id
		}

		return Message{id, event, data}, nil
	}

	return Message{}, es.err
}
