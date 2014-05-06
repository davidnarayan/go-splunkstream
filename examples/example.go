package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/davidnarayan/go-splunkstream"
)

func main() {
	//endpoint := "http://localhost:8089/services/receivers/stream?sourcetype=testevent&source=splunkstream"
	c, err := splunkstream.NewClient(
		"https://localhost:8089",
		"admin",
		"changeme",
		"splunkstream",
		"testevent",
	)

	if err != nil {
		log.Fatal(err)
	}

	// Use an id just to make things easier to find in Splunk
	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(1000)

	// Send events to Splunk
	n := 10
	t0 := time.Now()

	for i := 0; i < n; i++ {
		event := fmt.Sprintf("%s [stream_id=%03d] Test event %d\n", time.Now(), id, i)
		c.Send(event)
	}

	log.Printf("Sent %d events in %s", n, time.Now().Sub(t0))
}
