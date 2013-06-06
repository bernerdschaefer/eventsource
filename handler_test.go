package eventsource

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type testCloseNotifier struct {
	closed chan bool
	http.ResponseWriter
}

func (n testCloseNotifier) Close() {
	n.closed <- true
}

func (n testCloseNotifier) CloseNotify() <-chan bool {
	return n.closed
}

func TestHandlerValidatesAcceptHeader(t *testing.T) {
	handler := func(enc *Encoder, stop <-chan bool) {}

	w, r := httptest.NewRecorder(), &http.Request{}
	Handler(handler).ServeHTTP(w, r)

	if w.HeaderMap.Get("Content-Type") != "text/event-stream" {
		t.Fatal("handler did not set appropriate content type")
	}

	if w.Code != http.StatusNotAcceptable {
		t.Fatal("handler did not set 406 status")
	}
}

func TestHandlerSetsContentType(t *testing.T) {
	handler := func(enc *Encoder, stop <-chan bool) {}
	w, r := httptest.NewRecorder(), &http.Request{Header: map[string][]string{
		"Accept": []string{"text/event-stream"},
	}}
	Handler(handler).ServeHTTP(w, r)

	if w.HeaderMap.Get("Content-Type") != "text/event-stream" {
		t.Fatal("handler did not set appropriate content type")
	}

	if w.Code != http.StatusOK {
		t.Fatal("handler did not set 200 status")
	}
}

func TestHandlerEncode(t *testing.T) {
	handler := func(enc *Encoder, stop <-chan bool) {
		enc.Encode(Event{Data: []byte("hello")})
	}

	w, r := httptest.NewRecorder(), &http.Request{Header: map[string][]string{
		"Accept": []string{"text/event-stream"},
	}}

	Handler(handler).ServeHTTP(w, r)

	var event Event
	NewDecoder(w.Body).Decode(&event)

	if !reflect.DeepEqual(event, Event{Type: "message", Data: []byte("hello")}) {
		t.Error("unexpected handler output")
	}
}

func TestHandlerCloseNotify(t *testing.T) {
	done := make(chan bool, 1)
	handler := func(enc *Encoder, stop <-chan bool) {
		<-stop
		done <- true
	}

	w, r := httptest.NewRecorder(), &http.Request{Header: map[string][]string{
		"Accept": []string{"text/event-stream"},
	}}
	closer := testCloseNotifier{make(chan bool, 1), w}
	go Handler(handler).ServeHTTP(closer, r)

	closer.Close()
	select {
	case <-done:
	case <-time.After(time.Millisecond):
		t.Error("handler was not notified of close")
	}
}
