package eventsource_test

import (
	"fmt"
	"github.com/bernerdschaefer/eventsource"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func ExampleEncoder() {
	enc := eventsource.NewEncoder(os.Stdout)

	messages := []eventsource.Message{
		{ID: []byte("1"), Data: []byte("data")},
		{ID: []byte(""), Data: []byte("id reset")},
		{Event: []byte("add"), Data: []byte("1")},
	}

	for _, message := range messages {
		if err := enc.Write(message); err != nil {
			log.Fatal(err)
		}
	}

	if err := enc.WriteField("", []byte("heartbeat")); err != nil {
		log.Fatal(err)
	}

	if err := enc.Flush(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// id: 1
	// data: data
	//
	// id
	// data: id reset
	//
	// event: add
	// data: 1
	//
	// : heartbeat
	//
}

func ExampleDecoder() {
	stream := strings.NewReader(`id: 1
event: add
data: 123

id: 2
event: remove
data: 321

id: 3
event: add
data: 123

`)
	dec := eventsource.NewDecoder(stream)

	for {
		id, event, data, err := dec.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s. %s %s\n", id, event, data)
	}

	// Output:
	// 1. add 123
	// 2. remove 321
	// 3. add 123
}

func ExampleNew() {
	req, _ := http.NewRequest("GET", "http://localhost:9090/events", nil)
	req.SetBasicAuth("user", "pass")

	es := eventsource.New(req, 3*time.Second)

	for {
		message, err := es.Read()

		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%s. %s %s\n", message.ID, message.Event, message.Data)
	}
}
