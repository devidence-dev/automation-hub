// Package messages holds the user-facing texts of the hub's output. The defaults are
// embedded; the config's messages: section overrides single keys.
package messages

import (
	"bytes"
	_ "embed" // default.yaml
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/viper"
)

//go:embed default.yaml
var defaultYAML []byte

// placeholderPattern matches a named placeholder such as {date}.
var placeholderPattern = regexp.MustCompile(`\{[a-z_]+\}`)

// Catalog maps dotted keys, such as caption.header, to their texts.
type Catalog struct {
	texts map[string]string
}

// Load returns the embedded default catalog with the overrides applied. Overrides use
// the same nested keys as the defaults and replace only the keys they name.
func Load(overrides map[string]any) (*Catalog, error) {
	return load(defaultYAML, overrides)
}

func load(defaults []byte, overrides map[string]any) (*Catalog, error) {
	base, err := parseTexts(defaults)
	if err != nil {
		return nil, fmt.Errorf("parse default messages: %w", err)
	}
	merged := maps.Clone(base)
	if err := applyOverrides(merged, base, "", overrides); err != nil {
		return nil, err
	}
	return &Catalog{texts: merged}, nil
}

// Lookup returns the text of key and whether the catalog has it.
func (c *Catalog) Lookup(key string) (string, bool) {
	text, ok := c.texts[key]
	return text, ok
}

// Keys returns every key of the catalog, sorted.
func (c *Catalog) Keys() []string {
	return slices.Sorted(maps.Keys(c.texts))
}

// Render returns the text of key with each {name} placeholder replaced by vars[name].
// An unknown key, or a placeholder without a value, is an error.
func (c *Catalog) Render(key string, vars map[string]string) (string, error) {
	text, ok := c.texts[key]
	if !ok {
		return "", fmt.Errorf("unknown message key %q", key)
	}
	var missing []string
	out := placeholderPattern.ReplaceAllStringFunc(text, func(p string) string {
		value, ok := vars[p[1:len(p)-1]]
		if !ok {
			missing = append(missing, p)
			return p
		}
		return value
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("message %q has no value for %s", key, strings.Join(missing, ", "))
	}
	return out, nil
}

// parseTexts flattens a YAML document of nested texts into dotted keys.
func parseTexts(data []byte) (map[string]string, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	texts := map[string]string{}
	if err := flatten("", v.AllSettings(), texts); err != nil {
		return nil, err
	}
	return texts, nil
}

func flatten(prefix string, node map[string]any, texts map[string]string) error {
	for k, v := range node {
		path := join(prefix, k)
		switch val := v.(type) {
		case map[string]any:
			if err := flatten(path, val, texts); err != nil {
				return err
			}
		case string:
			texts[path] = val
		default:
			return fmt.Errorf("%s must be a text", path)
		}
	}
	return nil
}

// applyOverrides sets each override over its default text. An unknown key, a text where a
// group is expected, a map where a text is expected and a changed placeholder are errors.
func applyOverrides(merged, base map[string]string, prefix string, node map[string]any) error {
	for k, v := range node {
		path := join(prefix, k)
		def, isText := base[path]
		if !isText {
			if err := applyGroup(merged, base, path, v); err != nil {
				return err
			}
			continue
		}
		text, ok := v.(string)
		if !ok {
			return fmt.Errorf("messages: %s must be a text", path)
		}
		if err := checkPlaceholders(path, def, text); err != nil {
			return err
		}
		merged[path] = text
	}
	return nil
}

// applyGroup descends into a nested override. The path must be a group of texts in the defaults.
func applyGroup(merged, base map[string]string, path string, v any) error {
	if !hasGroup(base, path) {
		return fmt.Errorf("messages: unknown key %s", path)
	}
	sub, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("messages: %s is a group of texts, not a text", path)
	}
	return applyOverrides(merged, base, path, sub)
}

// checkPlaceholders requires the override to use exactly the placeholders of the default.
func checkPlaceholders(path, def, text string) error {
	want := placeholderPattern.FindAllString(def, -1)
	got := placeholderPattern.FindAllString(text, -1)
	for _, p := range want {
		if !slices.Contains(got, p) {
			return fmt.Errorf("messages: %s drops the placeholder %s", path, p)
		}
	}
	for _, p := range got {
		if !slices.Contains(want, p) {
			return fmt.Errorf("messages: %s adds the placeholder %s", path, p)
		}
	}
	return nil
}

// hasGroup reports whether some default key lies under path.
func hasGroup(texts map[string]string, path string) bool {
	for key := range texts {
		if strings.HasPrefix(key, path+".") {
			return true
		}
	}
	return false
}

func join(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
