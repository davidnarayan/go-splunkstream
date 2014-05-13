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

// Config holds everything required to connect and create events in Splunk
type Config struct {
	Scheme   string // http or https
	Host     string // host:port of the Splunk HTTP receiver
	Username string // username that has access to Splunk (requires edit_tcp)
	Password string // password for access to Splunk
	Endpoint string // API endpoint for the HTTP stream receiver

	// Query args for receiver
	Source     string // maps to "source" in Splunk
	SourceType string // maps to "sourcetype" in Splunk
	RemoteHost string // maps to "host" in Splunk
	Index      string // maps to "index" in Splunk
}

// SetDefaults sets the default configuration values for connecting to the
// HTTP stream receiver
func (cf *Config) SetDefaults() {
	if cf.Scheme == "" {
		cf.Scheme = "https"
	}

	if cf.Host == "" {
		cf.Host = "localhost:8089"
	}

	if cf.Username == "" {
		cf.Username = "admin"
	}

	if cf.Password == "" {
		cf.Password = "changeme"
	}

	if cf.Endpoint == "" {
		cf.Endpoint = "/services/receivers/stream"
	}

	if cf.Source == "" {
		cf.Source = "splunkstream.go"
	}

	if cf.SourceType == "" {
		cf.SourceType = "splunkstream"
	}
}

// URL builds the URL for connecting to the Splunk HTTP stream receiver
func (cf *Config) URL() string {
	u, q := cf.url()
	return u.String() + "?" + q
}

// RequestURI builds the URI path for the Splunk HTTP stream receiver. This is
// similar to URL() but does not have the scheme, host, or port
func (cf *Config) RequestURI() string {
	u, q := cf.url()
	return u.Path + "?" + q
}

// url builds the url for the HTTP stream receiver including the API anedpoint
// and any query args.
func (cf *Config) url() (*url.URL, string) {
	cf.SetDefaults()

	u := &url.URL{
		Scheme: cf.Scheme,
		Host:   cf.Host,
		Path:   cf.Endpoint,
	}

	v := url.Values{}
	v.Set("sourcetype", cf.SourceType)
	v.Set("source", cf.Source)

	if cf.Index != "" {
		v.Set("index", cf.Index)
	}

	if cf.RemoteHost != "" {
		v.Set("host", cf.RemoteHost)
	}

	return u, v.Encode()
}

//-----------------------------------------------------------------------------

// Client represents a client connection to Splunk
type Client struct {
	Config      *Config
	wroteHeader bool
	w           *bufio.Writer
	conn        net.Conn
}

// NewClient creates a new client connection to Splunk
func NewClient(config *Config) (*Client, error) {
	config.SetDefaults()
	var conn net.Conn

	if config.Scheme == "https" {
		// Splunk uses a self-signed certificate
		c, err := tls.Dial("tcp", config.Host, &tls.Config{InsecureSkipVerify: true})

		if err != nil {
			return nil, fmt.Errorf("Unable to create TLS connection to %s: %s",
				config.Host, err)
		}

		conn = c
	} else {
		c, err := net.Dial("tcp", config.Host)

		if err != nil {
			return nil, fmt.Errorf("Unable to create connection to %s: %s",
				config.Host, err)
		}

		conn = c
	}

	// Wrap the connection in a bufio Writer
	w := bufio.NewWriter(conn)

	return &Client{
		Config: config,
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
	fmt.Fprintf(c.w, "POST %s HTTP/1.1\r\n", c.Config.RequestURI())
	fmt.Fprintf(c.w, "Host: %s\r\n", c.Config.Host)
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
	return c.Config.URL()
}

// Close finishes a stream by flushing anything in the buffer to the receiver
func (c *Client) Close() {
	c.w.Flush()
	c.conn.Close()
}
