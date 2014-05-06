package splunkstream

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient(&Config{})

	if err != nil {
		t.Error(err)
	}
}
