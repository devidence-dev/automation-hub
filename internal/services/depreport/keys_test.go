package depreport

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"automation-hub/internal/messages"
)

var placeholderInText = regexp.MustCompile(`\{[a-z_]+\}`)

func defaultCatalog(t *testing.T) *messages.Catalog {
	t.Helper()
	cat, err := messages.Load(nil)
	if err != nil {
		t.Fatalf("messages.Load(nil) error = %v", err)
	}
	return cat
}

func TestDefaultsHaveEveryKey(t *testing.T) {
	cat := defaultCatalog(t)
	for _, key := range allKeys {
		if text, ok := cat.Lookup(key); !ok || text == "" {
			t.Errorf("default catalog has no text for %s", key)
		}
	}
	for _, key := range cat.Keys() {
		if !slices.Contains(allKeys, key) {
			t.Errorf("default catalog key %s has no constant", key)
		}
	}
}

func TestDefaultsKeepPlaceholders(t *testing.T) {
	cat := defaultCatalog(t)
	for _, key := range cat.Keys() {
		text, _ := cat.Lookup(key)
		for _, p := range placeholderInText.FindAllString(text, -1) {
			if !slices.Contains(placeholderNames, strings.Trim(p, "{}")) {
				t.Errorf("%s uses placeholder %s, which the code does not fill", key, p)
			}
		}
	}
}
