package media

import (
	"testing"
	"time"
)

func TestNewProvider(t *testing.T) {
	for _, tt := range []struct {
		name string
		fail bool
	}{{"omdb", false}, {"unknown", true}} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewProvider(tt.name, "http://localhost", "key", time.Second)
			if (err != nil) != tt.fail || (p == nil) != tt.fail {
				t.Fatalf("provider %v error %v", p, err)
			}
		})
	}
}
