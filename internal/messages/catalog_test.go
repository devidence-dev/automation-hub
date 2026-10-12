package messages

import (
	"strings"
	"testing"
)

func mustLoad(t *testing.T, overrides map[string]any) *Catalog {
	t.Helper()
	cat, err := Load(overrides)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return cat
}

func TestLoadNestedOverride(t *testing.T) {
	defaults := mustLoad(t, nil)
	cat := mustLoad(t, map[string]any{"caption": map[string]any{"header": "🔔 <b>Deps</b> · {date}"}})

	if got, _ := cat.Lookup("caption.header"); got != "🔔 <b>Deps</b> · {date}" {
		t.Errorf("caption.header = %q", got)
	}
	for _, key := range defaults.Keys() {
		if key == "caption.header" {
			continue
		}
		want, _ := defaults.Lookup(key)
		if got, _ := cat.Lookup(key); got != want {
			t.Errorf("%s = %q, want the default %q", key, got, want)
		}
	}
}

func TestLoadRejectsUnknownKey(t *testing.T) {
	_, err := Load(map[string]any{"caption": map[string]any{"headr": "x"}})
	if err == nil || !strings.Contains(err.Error(), "caption.headr") {
		t.Errorf("Load() error = %v, want one naming caption.headr", err)
	}
}

func TestLoadRejectsDroppedPlaceholder(t *testing.T) {
	_, err := Load(map[string]any{"caption": map[string]any{"more_repos": "… más"}})
	if err == nil || !strings.Contains(err.Error(), "caption.more_repos") || !strings.Contains(err.Error(), "{count}") {
		t.Errorf("Load() error = %v, want one naming caption.more_repos and {count}", err)
	}
}

func TestLoadRejectsAddedPlaceholder(t *testing.T) {
	_, err := Load(map[string]any{"caption": map[string]any{"errors": "⚠️ {count} {extra}"}})
	if err == nil || !strings.Contains(err.Error(), "{extra}") {
		t.Errorf("Load() error = %v, want one naming {extra}", err)
	}
}

func TestLoadRejectsMapWhereTextExpected(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]any
		wantKey   string
	}{
		{name: "map under a text", overrides: map[string]any{"fork": map[string]any{"tag_caption": map[string]any{"x": "y"}}}, wantKey: "fork.tag_caption"},
		{name: "text where a group is expected", overrides: map[string]any{"caption": "Dependency report"}, wantKey: "caption"},
		{name: "number where a text is expected", overrides: map[string]any{"lang": 5}, wantKey: "lang"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(tt.overrides)
			if err == nil || !strings.Contains(err.Error(), tt.wantKey) {
				t.Errorf("Load() error = %v, want one naming %s", err, tt.wantKey)
			}
		})
	}
}

func TestLoadRejectsInvalidDefaults(t *testing.T) {
	if _, err := load([]byte("a: [unclosed"), nil); err == nil {
		t.Error("load() accepted invalid YAML")
	}
	if _, err := load([]byte("a:\n  b: 1\n"), nil); err == nil || !strings.Contains(err.Error(), "a.b") {
		t.Errorf("load() error = %v, want one naming a.b", err)
	}
}

func TestRenderFillsPlaceholders(t *testing.T) {
	cat := mustLoad(t, nil)
	got, err := cat.Render("caption.header", map[string]string{"date": "2026-10-12"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if want := "📦 <b>Dependency report</b> · 2026-10-12"; got != want {
		t.Errorf("Render() = %q, want %q", got, want)
	}
}

func TestRenderMissingValueErrors(t *testing.T) {
	cat := mustLoad(t, nil)
	if _, err := cat.Render("caption.header", nil); err == nil || !strings.Contains(err.Error(), "{date}") {
		t.Errorf("Render() error = %v, want one naming {date}", err)
	}
}

func TestRenderUnknownKeyErrors(t *testing.T) {
	cat := mustLoad(t, nil)
	if _, err := cat.Render("caption.headr", nil); err == nil {
		t.Error("Render() accepted an unknown key")
	}
}

func TestRenderKeepsValuesWhole(t *testing.T) {
	// A value that looks like a placeholder is inserted as written, not filled again.
	cat := mustLoad(t, nil)
	got, err := cat.Render("caption.header", map[string]string{"date": "{count}"})
	if err != nil || got != "📦 <b>Dependency report</b> · {count}" {
		t.Errorf("Render() = %q, %v", got, err)
	}
}
