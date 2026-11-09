package media

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Metadata holds the fields provided for an identified title.
type Metadata struct{ Title, Year, IMDbID, IMDbRating, RottenTomatoes, Genre, Runtime, Rated, Poster string }

// Provider looks up metadata for a release.
type Provider interface {
	Lookup(context.Context, Release) (*Metadata, error)
}

// ErrNotFound indicates that the provider did not identify a title.
var ErrNotFound = errors.New("media not found")

// NewProvider constructs the configured metadata provider.
func NewProvider(name, baseURL, apiKey string, timeout time.Duration) (Provider, error) {
	switch name {
	case "omdb":
		return NewOMDb(baseURL, apiKey, timeout), nil
	default:
		return nil, fmt.Errorf("unknown metadata provider %q", name)
	}
}
