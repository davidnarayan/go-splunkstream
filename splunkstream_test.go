package splunkstream

import (
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"
    "os"
    "path/filepath"
)

// These flags allow the Splunk connection information to be specified as
// command line options
var host = flag.String("splunk.host", "", "Splunk host:port")
var auth = flag.String("splunk.auth", "", "Splunk auth credentials")
var insecure = flag.Bool("splunk.insecure", false, "Use HTTP instead of HTTPS")

func config() *Config {
	cf := &Config{
        Host: *host,
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

	return cf
}
func TestNewClient(t *testing.T) {
	cf := config()
	_, err := NewClient(cf)

	if err != nil {
		t.Error(err)
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
