package eventsource_test

import (
	"fmt"
	"github.com/bernerdschaefer/eventsource"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

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
