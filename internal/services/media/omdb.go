package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"automation-hub/internal/httpx"
)

type omdbClient struct {
	baseURL, apiKey string
	client          *http.Client
	posterWidth     int
}

// omdbRating is one entry of the OMDb Ratings list.
type omdbRating struct{ Source, Value string }

// NewOMDb constructs a client with a whole-request timeout. posterWidth is the poster width to request, 0 keeps the original.
func NewOMDb(baseURL, apiKey string, timeout time.Duration, posterWidth int) Provider {
	return &omdbClient{baseURL: baseURL, apiKey: apiKey, client: httpx.NewClient(httpx.Settings{Timeout: timeout}), posterWidth: posterWidth}
}
func (o *omdbClient) Lookup(ctx context.Context, release Release) (*Metadata, error) {
	u, err := url.Parse(o.baseURL)
	if err != nil {
		return nil, errors.New("invalid metadata provider URL")
	}
	q := u.Query()
	q.Set("t", release.Title)
	q.Set("apikey", o.apiKey)
	if release.Year != "" {
		q.Set("y", release.Year)
	}
	u.RawQuery = q.Encode()
	resp, err := httpx.Send(ctx, o.client, httpx.Request{Method: http.MethodGet, URL: u.String()}, httpx.DefaultMaxResponseBytes)
	if err != nil {
		return nil, errors.New("metadata provider request failed")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metadata provider returned HTTP %d", resp.StatusCode)
	}
	var data struct {
		Title, Year, IMDbID, IMDbRating, IMDbVotes, Genre, Runtime, Rated, Poster, Response string
		Director, Actors, Plot                                                              string
		Ratings                                                                             []omdbRating
	}
	if json.Unmarshal(resp.Body, &data) != nil {
		return nil, errors.New("invalid metadata provider response")
	}
	if data.Response != "True" {
		return nil, ErrNotFound
	}
	m := &Metadata{Title: clean(data.Title), Year: clean(data.Year), IMDbID: clean(data.IMDbID), IMDbRating: clean(data.IMDbRating), IMDbVotes: clean(data.IMDbVotes), Genre: clean(data.Genre), Runtime: clean(data.Runtime), Rated: clean(data.Rated), Director: clean(data.Director), Actors: clean(data.Actors), Plot: clean(data.Plot)}
	m.Poster, m.PosterFallback = o.resolvePoster(clean(data.Poster))
	m.RottenTomatoes = ratingValue(data.Ratings, "Rotten Tomatoes")
	m.Metacritic = ratingValue(data.Ratings, "Metacritic")
	return m, nil
}

// resolvePoster resizes the poster to the configured width. The fallback is empty when the URL did not change.
func (o *omdbClient) resolvePoster(poster string) (string, string) {
	resized, original := ResizePoster(poster, o.posterWidth)
	if resized == original {
		return resized, ""
	}
	return resized, original
}

// ratingValue returns the cleaned value of the rating from source, or empty when OMDb lists none.
func ratingValue(ratings []omdbRating, source string) string {
	for _, rating := range ratings {
		if rating.Source == source {
			return clean(rating.Value)
		}
	}
	return ""
}

// clean turns the OMDb placeholder N/A into an empty value.
func clean(s string) string {
	if s == "N/A" {
		return ""
	}
	return s
}
