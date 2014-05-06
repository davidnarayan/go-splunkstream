// Package splunkstream is a library for the Splunk HTTP streaming API
package splunkstream

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/url"
)

//-----------------------------------------------------------------------------

var defaultPath = "/services/receivers/stream"

type Client struct {
	conn       net.Conn
	url        *url.URL
	username   string
	password   string
	sentHeader bool

	bw *bufio.Writer
}

func NewClient(endpoint, username, password, source, sourcetype string) (*Client, error) {
	rawurl := fmt.Sprintf("%s%s?sourcetype=%s&source=%s", endpoint, defaultPath,
		sourcetype, source)
	u, err := url.Parse(rawurl)

	if err != nil {
		return nil, err
	}

	var conn net.Conn

	if u.Scheme == "https" {
		// Splunk uses a self-signed certificate
		c, err := tls.Dial("tcp", u.Host, &tls.Config{InsecureSkipVerify: true})

		if err != nil {
			return nil, fmt.Errorf("Unable to create TLS connection to %s: %s",
				u, err)
		}

		conn = c
	} else {
		c, err := net.Dial("tcp", u.Host)

		if err != nil {
			return nil, fmt.Errorf("Unable to create connection to %s: %s",
				u, err)
		}

		conn = c
	}

	// Wrap the connection in a bufio Writer
	bw := bufio.NewWriter(conn)

	return &Client{
		conn:     conn,
		url:      u,
		bw:       bw,
		username: username,
		password: password,
	}, nil
}

// basicAuth returns base64 encoded credentials as per RFC2617. Also see
// basicAuth in net/http/client.go.
func (c *Client) basicAuth() string {
	auth := c.username + ":" + c.password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// sendHeader sends the initial POST request with Authorization and HTTP
// headers.
func (c *Client) sendHeader() {
	fmt.Fprintf(c.bw, "POST %s HTTP/1.1\r\n", c.url.RequestURI())
	fmt.Fprintf(c.bw, "Host: %s\r\n", c.url.Host)
	fmt.Fprintf(c.bw, "Authorization: Basic %s\r\n", c.basicAuth())
	io.WriteString(c.bw, "x-splunk-input-mode: streaming\r\n")
	io.WriteString(c.bw, "\r\n")

	c.sentHeader = true
}

// Send streams an event to Splunk
func (c *Client) Send(s string) {
	if !c.sentHeader {
		c.sendHeader()
	}

	io.WriteString(c.bw, s)
	io.WriteString(c.bw, "\r\n")
}
