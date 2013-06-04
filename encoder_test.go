package eventsource

import (
	"bytes"
	"testing"
)

func TestWriteField(t *testing.T) {
	table := []struct {
		field string
		value []byte
		out   string
		error
	}{
		{"data", []byte("data"), "data: data\n", nil},
		{"data", nil, "data\n", nil},
		{"\xFF\xFE\xFD", nil, "", InvalidEncodingErr},
		{"data", []byte("\xFF\xFE\xFD"), "", InvalidEncodingErr},
		{"data", []byte("a\nb\nc\n"), "data: a\ndata: b\ndata: c\ndata\n", nil},
		{"data", []byte("a\r\nb\r\nc"), "data: a\ndata: b\ndata: c\n", nil},
	}

	for i, tt := range table {
		buf := new(bytes.Buffer)

		err := NewEncoder(buf).WriteField(tt.field, tt.value)

		if tt.error != nil && err == tt.error {
			continue
		}

		if tt.error != err {
			t.Errorf("%d. expected err=%q, got %q", i, tt.error, err)
			continue
		}

		if buf.String() != tt.out {
			t.Errorf("%d. expected %q, got %q", i, tt.out, buf.String())
		}
	}
}

func TestEncoderEncode(t *testing.T) {
	table := []struct {
		Event
		expected string
	}{
		{Event{Type: "type"}, "event: type\ndata\n\n"},
		{Event{ID: "123"}, "id: 123\ndata\n\n"},
		{Event{Retry: "10000"}, "retry: 10000\ndata\n\n"},
		{Event{Data: []byte("data")}, "data: data\n\n"},
		{Event{ResetID: true}, "id\ndata\n\n"},
	}

	for i, tt := range table {
		buf := new(bytes.Buffer)

		if err := NewEncoder(buf).Encode(tt.Event); err != nil {
			t.Errorf("%d. write error: %q", i, err)
			continue
		}

		if buf.String() != tt.expected {
			t.Errorf("%d. expected %q, got %q", i, tt.expected, buf.String())
		}
	}
}
