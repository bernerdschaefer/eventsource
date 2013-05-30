package eventsource

import (
	"fmt"
	"io"
)

// The FlushWriter interface groups basic Write and Flush methods.
type FlushWriter interface {
	io.Writer
	Flush()
}

// Adds a noop Flush method to a normal io.Writer.
type noopFlusher struct {
	io.Writer
}

func (noopFlusher) Flush() {}

// Encoder writes EventSource messages to an output stream.
type Encoder struct {
	w FlushWriter
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	if w, ok := w.(FlushWriter); ok {
		return &Encoder{w}
	}

	return &Encoder{noopFlusher{w}}
}

// Flush sends an empty line to signal message is complete, and flushes the
// writer.
func (e *Encoder) Flush() error {
	_, err := e.w.Write([]byte{'\n'})
	return err
}

// WriteField writes a message field to the connection. If value is nil, this
// is a noop.
func (e *Encoder) WriteField(field string, value []byte) (err error) {
	if value == nil {
		// nothing to do
		return
	}

	if len(value) == 0 {
		_, err = fmt.Fprintf(e.w, "%s\n", field)
	} else {
		_, err = fmt.Fprintf(e.w, "%s: %s\n", field, value)
	}

	return
}

// Write wries a message to the connection. If m contains no data, this is a
// noop.
func (e *Encoder) Write(m Message) error {
	if len(m.Data) == 0 {
		return nil
	}

	if err := e.WriteField("id", m.ID); err != nil {
		return err
	}

	if err := e.WriteField("event", m.Event); err != nil {
		return err
	}

	if err := e.WriteField("data", m.Data); err != nil {
		return err
	}

	return e.Flush()
}
