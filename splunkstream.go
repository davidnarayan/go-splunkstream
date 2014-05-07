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

type Config struct {
	Host       string
	Username   string
	Password   string
	Source     string
	SourceType string
	Scheme     string
	Endpoint   string
}

func (cf *Config) setDefaults() {
	if cf.Host == "" {
		cf.Host = "localhost:8089"
	}

	if cf.Username == "" {
		cf.Username = "admin"
	}

	if cf.Password == "" {
		cf.Password = "changeme"
	}

	if cf.Source == "" {
		cf.Source = "splunkstream.go"
	}

	if cf.SourceType == "" {
		cf.SourceType = "splunkstream"
	}

	if cf.Scheme == "" {
		cf.Scheme = "https"
	}

	if cf.Endpoint == "" {
		cf.Endpoint = "/services/receivers/stream"
	}
}

//-----------------------------------------------------------------------------

type Client struct {
	Config      *Config
	url         *url.URL
	wroteHeader bool
	w           *bufio.Writer
	conn        net.Conn
}

func NewClient(config *Config) (*Client, error) {
	config.setDefaults()

	rawurl := fmt.Sprintf("%s://%s%s?sourcetype=%s&source=%s", config.Scheme,
		config.Host, config.Endpoint, config.SourceType, config.Source)
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
	w := bufio.NewWriter(conn)

	return &Client{
		Config: config,
		url:    u,
		w:      w,
		conn:   conn,
	}, nil
}

// basicAuth returns base64 encoded credentials as per RFC2617. Also see
// basicAuth in net/http/client.go.
func (c *Client) basicAuth() string {
	auth := c.Config.Username + ":" + c.Config.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// writeHeader sends the initial POST request with Authorization and HTTP
// headers.
func (c *Client) writeHeader() {
	fmt.Fprintf(c.w, "POST %s HTTP/1.1\r\n", c.url.RequestURI())
	fmt.Fprintf(c.w, "Host: %s\r\n", c.url.Host)
	fmt.Fprintf(c.w, "Authorization: Basic %s\r\n", c.basicAuth())
	io.WriteString(c.w, "x-splunk-input-mode: streaming\r\n")
	io.WriteString(c.w, "\r\n")
	c.w.Flush()

	c.wroteHeader = true
}

// Write sends data to Splunk
func (c *Client) Write(b []byte) (n int, err error) {
	if !c.wroteHeader {
		c.writeHeader()
	}

	return c.w.Write(b)
}

// String returns a string representation of the splunkstream client as an
// HTTP endpoint
func (c *Client) String() string {
	return c.url.String()
}

// Close finishes a stream by flushing anything in the buffer to the receiver
func (c *Client) Close() {
	c.w.Flush()
	c.conn.Close()
}
