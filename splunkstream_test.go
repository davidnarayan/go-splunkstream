package splunkstream

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient("localhost:8089", &Config{})

	if err != nil {
		t.Error(err)
	}
}
