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
}

// NewOMDb constructs a client with a whole-request timeout.
func NewOMDb(baseURL, apiKey string, timeout time.Duration) Provider {
	return &omdbClient{baseURL: baseURL, apiKey: apiKey, client: httpx.NewClient(httpx.Settings{Timeout: timeout})}
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
		Title, Year, IMDbID, IMDbRating, Genre, Runtime, Rated, Poster, Response string
		Ratings                                                                  []struct{ Source, Value string }
	}
	if json.Unmarshal(resp.Body, &data) != nil {
		return nil, errors.New("invalid metadata provider response")
	}
	if data.Response != "True" {
		return nil, ErrNotFound
	}
	clean := func(s string) string {
		if s == "N/A" {
			return ""
		}
		return s
	}
	m := &Metadata{Title: clean(data.Title), Year: clean(data.Year), IMDbID: clean(data.IMDbID), IMDbRating: clean(data.IMDbRating), Genre: clean(data.Genre), Runtime: clean(data.Runtime), Rated: clean(data.Rated), Poster: clean(data.Poster)}
	for _, rating := range data.Ratings {
		if rating.Source == "Rotten Tomatoes" {
			m.RottenTomatoes = clean(rating.Value)
			break
		}
	}
	return m, nil
}
