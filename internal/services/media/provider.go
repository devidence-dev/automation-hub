package media

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Metadata holds the fields provided for an identified title.
// PosterFallback is the original poster URL when Poster was resized, empty otherwise.
type Metadata struct {
	Title, Year, IMDbID, IMDbRating, IMDbVotes, RottenTomatoes, Metacritic string
	Genre, Runtime, Rated, Director, Actors, Plot, Poster, PosterFallback  string
}

// Provider looks up metadata for a release.
type Provider interface {
	Lookup(context.Context, Release) (*Metadata, error)
}

// ErrNotFound indicates that the provider did not identify a title.
var ErrNotFound = errors.New("media not found")

// NewProvider constructs the configured metadata provider. posterWidth is the poster width to request, 0 keeps the original.
func NewProvider(name, baseURL, apiKey string, timeout time.Duration, posterWidth int) (Provider, error) {
	switch name {
	case "omdb":
		return NewOMDb(baseURL, apiKey, timeout, posterWidth), nil
	default:
		return nil, fmt.Errorf("unknown metadata provider %q", name)
	}
}
