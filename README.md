go-splunkstream
===============

go-splunkstream is a library for the Splunk HTTP streaming receiver

Quick Start
-----------

`go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/davidnarayan/go-splunkstream"
)

func main() {
	c, err := splunkstream.NewClient(
		&splunkstream.Config{
			Host:       "localhost:8089",
			Username:   "admin",
			Password:   "changeme",
			SourceType: "testevent",
			Source:     "splunkstream/example.go",
			Scheme:     "http",
		})

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
		c.Write([]byte(event))
	}

	c.Close()
	log.Printf("Sent %d events to %s in %s", n, c, time.Now().Sub(t0))
}
`
