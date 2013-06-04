package eventsource

import (
	"bytes"
	"testing"
)

func TestEncoderWrite(t *testing.T) {
	table := []struct {
		Event
		expected string
	}{
		{Event{}, ""},
		{Event{nil, nil, []byte("message")}, "data: message\n\n"},
		{Event{nil, []byte("add"), []byte("1")}, "event: add\ndata: 1\n\n"},
		{Event{[]byte("1"), []byte("add"), []byte("1")}, "id: 1\nevent: add\ndata: 1\n\n"},
		{Event{[]byte(""), []byte("add"), []byte("1")}, "id\nevent: add\ndata: 1\n\n"},
	}

	for i, tt := range table {
		buf := new(bytes.Buffer)

		if err := NewEncoder(buf).Write(tt.Event); err != nil {
			t.Errorf("%d. write error: %q", i, err)
			continue
		}

		if buf.String() != tt.expected {
			t.Errorf("%d. expected %q, got %q", i, tt.expected, buf.String())
		}
	}
}
