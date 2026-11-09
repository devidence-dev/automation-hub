package config

import (
	"strings"
	"testing"
	"time"
)

// metadataConfigCase is one hook metadata setting and the key that its error must name, or empty when valid.
type metadataConfigCase struct {
	name string
	m    *MetadataConfig
	want string
}

func TestMetadataConfig(t *testing.T) {
	for _, tt := range []metadataConfigCase{{"absent", nil, ""}, {"missing key", &MetadataConfig{Provider: "omdb"}, "api_key"}, {"provider", &MetadataConfig{Provider: "imdb", APIKey: "key"}, "provider"}, {"timeout", &MetadataConfig{Provider: "omdb", APIKey: "key", Timeout: -time.Second}, "timeout"}, {"defaults", &MetadataConfig{Provider: "omdb", APIKey: "key"}, ""}} {
		t.Run(tt.name, func(t *testing.T) {
			h := []WebhookConfig{{Name: "qbittorrent", Config: WebhookProcessorConfig{Metadata: tt.m}}}
			checkMetadataConfig(t, tt, validateHooks(h))
		})
	}
}

// checkMetadataConfig asserts that err names the expected key, or that a valid setting got its defaults.
func checkMetadataConfig(t *testing.T, tt metadataConfigCase, err error) {
	t.Helper()
	if tt.want != "" {
		if err == nil || !strings.Contains(err.Error(), "hook.qbittorrent.config.metadata."+tt.want) {
			t.Fatalf("error %v", err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if tt.m != nil && (tt.m.Timeout != DefaultMetadataTimeout || tt.m.BaseURL != DefaultMetadataBaseURL) {
		t.Fatalf("defaults %#v", tt.m)
	}
}
