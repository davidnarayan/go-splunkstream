package splunkstream

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// These flags allow the Splunk connection information to be specified as
// command line options
var host = flag.String("splunk.host", "", "Splunk host:port")
var auth = flag.String("splunk.auth", "", "Splunk auth credentials")
var insecure = flag.Bool("splunk.insecure", false, "Use HTTP instead of HTTPS")

func config() *Config {
	cf := &Config{
		Host:   *host,
		Source: filepath.Base(os.Args[0]),
	}

	s := strings.SplitN(*auth, ":", 2)

	if len(s) > 2 {
		cf.Username = s[0]
		cf.Password = s[1]
	}

	if *insecure {
		cf.Scheme = "http"
	}

	cf.SetDefaults()

	return cf
}

func TestNewClient(t *testing.T) {
	cf := config()
	_, err := NewClient(cf)

	if err != nil {
		t.Error(err)
	}
}

func TestURL(t *testing.T) {
	cf := config()
	cf.Index = "testindex"
	cf.RemoteHost = "testhost"
	cf.Source = "testsource"
	cf.SourceType = "testsourcetype"

	s := fmt.Sprintf("%s://%s%s?host=%s&index=%s&source=%s&sourcetype=%s",
		cf.Scheme, cf.Host, cf.Endpoint, cf.RemoteHost, cf.Index, cf.Source,
		cf.SourceType)

	u1, err := url.Parse(s)

	if err != nil {
		t.Error(err)
	}

	u2 := cf.URL()

	if u1.String() != u2 {
		t.Errorf("Invalid URL:\nwant:\t%s\ngot:\t%s\n", u1, u2)
	}
}

func TestRequestURI(t *testing.T) {
	cf := config()
	cf.Index = "testindex"
	cf.RemoteHost = "testhost"
	cf.Source = "testsource"
	cf.SourceType = "testsourcetype"

	s := fmt.Sprintf("%s?host=%s&index=%s&source=%s&sourcetype=%s",
		cf.Endpoint, cf.RemoteHost, cf.Index, cf.Source,
		cf.SourceType)

	u1, err := url.Parse(s)

	if err != nil {
		t.Error(err)
	}

	u2 := cf.RequestURI()

	if u1.String() != u2 {
		t.Errorf("Invalid URL:\nwant:\t%s\ngot:\t%s\n", u1, u2)
	}
}

func TestWrite(t *testing.T) {
	cf := config()
	c, err := NewClient(cf)

	if err != nil {
		t.Error(err)
	}

	defer c.Close()

	msg := fmt.Sprintf("%s Test event\n", time.Now())
	n, err := c.Write([]byte(msg))

	if err != nil {
		t.Error(err)
	}

	if n != len(msg) {
		t.Errorf("Write failed: want: %d\ngot: %d\n", len(msg), n)
	}
}
