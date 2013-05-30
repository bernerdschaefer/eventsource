package eventsource

import (
	"bytes"
	"io"
	"testing"
)

func longLine() string {
	buf := make([]byte, 4096)
	for i := 0; i < len(buf); i++ {
		buf[i] = 'a'
	}

	return string(buf)
}

type messageExpectation struct {
	id    []byte
	event []byte
	data  []byte
	error
}

var fixtures = []struct {
	source       string
	expectations []messageExpectation
}{
	{
		"data: message 1\r\n\r\ndata: message\r\ndata:2\r\n\r\ndata: message 3\r\n\r\n",
		[]messageExpectation{
			{nil, nil, []byte("message 1"), nil},
			{nil, nil, []byte("message\n2"), nil},
			{nil, nil, []byte("message 3"), nil},
			{nil, nil, nil, io.EOF},
		},
	},

	{
		": this is a comment\r\n\r\ndata: space\r\n\r\ndata:nospace\r\n\r\n",
		[]messageExpectation{
			{nil, nil, nil, nil},
			{nil, nil, []byte("space"), nil},
			{nil, nil, []byte("nospace"), nil},
		},
	},

	{
		"event: add\ndata: 123\n\nevent: remove\ndata: 321\n\nevent: add\ndata: 123\n\n",
		[]messageExpectation{
			{nil, []byte("add"), []byte("123"), nil},
			{nil, []byte("remove"), []byte("321"), nil},
			{nil, []byte("add"), []byte("123"), nil},
		},
	},

	{
		"data: first event\nid: 1\n\ndata:second event\nid\n\ndata: third event\n\n",
		[]messageExpectation{
			{[]byte("1"), nil, []byte("first event"), nil},
			{[]byte{}, nil, []byte("second event"), nil},
			{nil, nil, []byte("third event"), nil},
		},
	},

	{
		"data\n\ndata:\ndata:\n\ndata:\n",
		[]messageExpectation{
			{nil, nil, []byte(""), nil},
			{nil, nil, []byte("\n"), nil},
			{nil, nil, nil, io.EOF},
		},
	},

	{
		"data:"+longLine()+"\n\n",
		[]messageExpectation{
			{nil, nil, []byte(longLine()), nil},
		},
	},

}

func TestDecoder(t *testing.T) {
	for i, f := range fixtures {
		dec := NewDecoder(bytes.NewBufferString(f.source))

		for j, e := range f.expectations {
			id, event, data, err := dec.Read()

			if err != e.error {
				t.Errorf("%d.%d expected err=%q, got %q", i, j, e.error, err)
				continue
			}

			if e.id == nil && id != nil {
				t.Errorf("%d.%d expected id=nil, got %q", i, j, id)
			}

			if e.id != nil && id == nil {
				t.Errorf("%d.%d expected id==%q, got nil", i, j, e.id)
			}

			if !bytes.Equal(e.id, id) {
				t.Errorf("%d.%d expected id=%q, got %q", i, j, e.id, id)
			}

			if !bytes.Equal(e.event, event) {
				t.Errorf("%d.%d expected event=%q, got %q", i, j, e.event, event)
			}

			if e.event == nil && event != nil {
				t.Errorf("%d.%d expected event=nil, got %q", i, j, event)
			}

			if !bytes.Equal(e.data, data) {
				t.Errorf("%d.%d expected data=%q, got %q", i, j, e.data, data)
			}

			if e.data == nil && data != nil {
				t.Errorf("%d.%d expected data=nil, got %q", i, j, data)
			}

			if e.data != nil && data == nil {
				t.Errorf("%d.%d expected data=%q, got nil", i, j, e.data)
			}
		}

		if _, _, _, err := dec.Read(); err != io.EOF {
			t.Errorf("%d. expected last read to be EOF, was %s", i, err)
		}
	}
}
