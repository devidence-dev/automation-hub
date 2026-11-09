package config

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

// metadataConfigCase is one hook metadata setting and the key that its error must name, or empty when valid.
// For a valid setting, wantWidth and wantPlot are the effective poster width and plot length.
type metadataConfigCase struct {
	name                string
	m                   *MetadataConfig
	want                string
	wantWidth, wantPlot int
}

// metadataCfg returns a valid omdb setting with the given poster width and plot length, nil for absent.
func metadataCfg(width, plot *int) *MetadataConfig {
	return &MetadataConfig{Provider: "omdb", APIKey: "key", PosterWidth: width, PlotMaxChars: plot}
}

func TestMetadataConfig(t *testing.T) {
	for _, tt := range []metadataConfigCase{
		{name: "absent", m: nil},
		{name: "missing key", m: &MetadataConfig{Provider: "omdb"}, want: "api_key"},
		{name: "provider", m: &MetadataConfig{Provider: "imdb", APIKey: "key"}, want: "provider"},
		{name: "unknown provider with settings", m: &MetadataConfig{Provider: "imdb", APIKey: "key", PosterWidth: intPtr(0)}, want: "provider"},
		{name: "timeout", m: &MetadataConfig{Provider: "omdb", APIKey: "key", Timeout: -time.Second}, want: "timeout"},
		{name: "defaults", m: metadataCfg(nil, nil), wantWidth: DefaultMetadataPosterWidth, wantPlot: DefaultMetadataPlotMaxChars},
		{name: "explicit zero", m: metadataCfg(intPtr(0), intPtr(0))},
		{name: "explicit values", m: metadataCfg(intPtr(500), intPtr(80)), wantWidth: 500, wantPlot: 80},
		{name: "negative width", m: metadataCfg(intPtr(-1), nil), want: "poster_width"},
		{name: "negative plot length", m: metadataCfg(nil, intPtr(-1)), want: "plot_max_chars"},
	} {
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
	if tt.m != nil && (*tt.m.PosterWidth != tt.wantWidth || *tt.m.PlotMaxChars != tt.wantPlot) {
		t.Fatalf("poster width %d plot length %d, want %d and %d", *tt.m.PosterWidth, *tt.m.PlotMaxChars, tt.wantWidth, tt.wantPlot)
	}
}

// metadataDecodeCase is the poster width and plot length as YAML would set them; nil means the key is absent.
type metadataDecodeCase struct {
	name                string
	width, plot         *int
	wantWidth, wantPlot *int
}

// TestMetadataDecode checks that an absent key stays nil and an explicit 0 stays a pointer to 0.
func TestMetadataDecode(t *testing.T) {
	for _, tt := range []metadataDecodeCase{
		{name: "absent", wantWidth: nil, wantPlot: nil},
		{name: "explicit zero", width: intPtr(0), plot: intPtr(0), wantWidth: intPtr(0), wantPlot: intPtr(0)},
		{name: "explicit value", width: intPtr(500), plot: intPtr(80), wantWidth: intPtr(500), wantPlot: intPtr(80)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := decodeMetadata(t, tt)
			if !reflect.DeepEqual(m.PosterWidth, tt.wantWidth) || !reflect.DeepEqual(m.PlotMaxChars, tt.wantPlot) {
				t.Fatalf("decoded width %v plot %v, want %v and %v", m.PosterWidth, m.PlotMaxChars, tt.wantWidth, tt.wantPlot)
			}
		})
	}
}

// decodeMetadata decodes the case's keys into a MetadataConfig with viper, as Load does.
func decodeMetadata(t *testing.T, tt metadataDecodeCase) MetadataConfig {
	t.Helper()
	v := viper.New()
	if tt.width != nil {
		v.Set("poster_width", *tt.width)
	}
	if tt.plot != nil {
		v.Set("plot_max_chars", *tt.plot)
	}
	var m MetadataConfig
	if err := v.Unmarshal(&m); err != nil {
		t.Fatal(err)
	}
	return m
}
