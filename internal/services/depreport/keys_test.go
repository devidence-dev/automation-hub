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

// ownKeys returns the catalog keys this package shows. The bot keys belong to the handlers.
func ownKeys(cat *messages.Catalog) []string {
	var keys []string
	for _, key := range cat.Keys() {
		if !strings.HasPrefix(key, "bot.") {
			keys = append(keys, key)
		}
	}
	return keys
}

func TestDefaultsHaveEveryKey(t *testing.T) {
	cat := defaultCatalog(t)
	for _, key := range allKeys {
		if text, ok := cat.Lookup(key); !ok || text == "" {
			t.Errorf("default catalog has no text for %s", key)
		}
	}
	for _, key := range ownKeys(cat) {
		if !slices.Contains(allKeys, key) {
			t.Errorf("default catalog key %s has no constant", key)
		}
	}
}

func TestDefaultsKeepPlaceholders(t *testing.T) {
	cat := defaultCatalog(t)
	for _, key := range ownKeys(cat) {
		text, _ := cat.Lookup(key)
		for _, p := range placeholderInText.FindAllString(text, -1) {
			if !slices.Contains(placeholderNames, strings.Trim(p, "{}")) {
				t.Errorf("%s uses placeholder %s, which the code does not fill", key, p)
			}
		}
	}
}

// TestOverrideSummaryEmojiUnused checks that the emoji keys stay valid for existing overrides, and that
// the document shows no emoji for a repo, so the override changes nothing on the page.
func TestOverrideSummaryEmojiUnused(t *testing.T) {
	overrides := map[string]any{"document": map[string]any{"summary": map[string]any{"emoji": map[string]any{"major": "🔥"}}}}
	out := presentDocument(t, overrides, summaryAt(reportOf(t, repoJSON("devidence-dev/QuantWarden", "major", "patch"))))
	if strings.Contains(out, "🔥") {
		t.Error("document renders the unused summary emoji override")
	}
	if !strings.Contains(out, "tone-major") {
		t.Error("document has no major dot for a major update")
	}
}

func TestUnusedKeyOverrideValidates(t *testing.T) {
	overrides := map[string]any{"document": map[string]any{"columns": map[string]any{"dependency": "Paquete"}}}
	if _, err := messages.Load(overrides); err != nil {
		t.Fatalf("messages.Load() error = %v", err)
	}
	// The table headers of the current document still render the column texts; the component
	// document does not. Only the startup validation is checked here.
	if out := presentDocument(t, overrides, summaryAt(loadFixtureReport(t))); out == "" {
		t.Error("document is empty after an unused key override")
	}
}

// TestNewKeysRender checks that each key added by the liquid glass port renders with its placeholders
// filled, so a missing or misspelled placeholder fails here and not in a document.
func TestNewKeysRender(t *testing.T) {
	cat := defaultCatalog(t)
	vars := map[string]string{"total": "53", "up_to_date": "1", "updates": "52", "deps": "52", "count": "3", "where": "raíz", "msg": "boom"}
	for _, key := range []string{
		KeyDocPanelSub, KeyDocStatusUpdatesBold1, KeyDocStatusUpdatesRest1, KeyDocStatusImagesRest, KeyDocStatusImagesRest1,
		KeyDocStatusImagesUpTo, KeyDocTypePatch, KeyDocTypeMinor, KeyDocTypeMajor, KeyDocThemeGroup, KeyDocThemeSystem,
		KeyDocThemeLight, KeyDocThemeDark, KeyDocEyebrowDeps, KeyDocEyebrowImages, KeyErrorsCount, KeyErrorsCountOne,
	} {
		out, err := cat.Render(key, vars)
		if err != nil {
			t.Errorf("Render(%s) error = %v", key, err)
			continue
		}
		if placeholderInText.MatchString(out) {
			t.Errorf("%s still has a placeholder after rendering: %q", key, out)
		}
	}
}

// TestUnusedBatch3KeysStillValidate checks that the keys the liquid glass port stopped rendering still pass
// startup validation with an override, so an existing messages: section keeps loading.
func TestUnusedBatch3KeysStillValidate(t *testing.T) {
	overrides := map[string]any{
		"document": map[string]any{
			"brand":   "Reporte",
			"total":   "Todos",
			"errors":  map[string]any{"item": "{where}: {msg}"},
			"summary": map[string]any{"counter": "{emoji} {count}"},
		},
	}
	if _, err := messages.Load(overrides); err != nil {
		t.Fatalf("messages.Load() error = %v", err)
	}
}
