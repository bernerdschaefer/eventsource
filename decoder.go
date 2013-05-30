package eventsource

import (
	"bufio"
	"bytes"
	"io"
)

// A decoder reads and decodes EventSource messages from an input stream.
type Decoder struct {
	r *bufio.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

// Reads an event. If a returned []byte value is nil, it was not included in
// the event.
func (d *Decoder) Read() (id, event, data []byte, err error) {
	var buf bytes.Buffer
	var dataBuf bytes.Buffer

	for {
		// BUG(bernerd): The EventSource spec defines valid line endings as CR, LF,
		// CRLF. We only support LF or CRLF.
		l, isPrefix, err := d.r.ReadLine()

		if err != nil {
			return nil, nil, nil, err
		}

		buf.Write(l)

		if isPrefix {
			continue
		}

		line := make([]byte, buf.Len())
		copy(line, buf.Bytes())
		buf.Reset()

		if len(line) == 0 {
			break
		}

		if line[0] == ':' {
			continue // comment
		}

		var field string
		var value []byte
		parts := bytes.SplitN(line, []byte{':'}, 2)

		if len(parts) == 2 {
			field = string(parts[0])
			value = parts[1]
		} else {
			field = string(parts[0])
			value = []byte{}
		}

		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}

		switch field {
		// BUG(bernerd): Server sent "retry" fields are ignored. Should this be
		// supported?
		case "id":
			id = value
		case "event":
			event = value
		case "data":
			dataBuf.Write(value)
			dataBuf.WriteRune('\n')
		}
	}

	data = dataBuf.Bytes()

	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	return
}
