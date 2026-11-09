package media

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	omdbAmazonPoster = "https://m.media-amazon.com/images/M/MV5BNzk@._V1_SX300.jpg"
	omdbFoundBody    = `{"Response":"True","Title":"Coyote vs. Acme","Year":"2021","imdbID":"tt1756855","imdbRating":"7.6","imdbVotes":"12,345",` +
		`"Genre":"Animation","Runtime":"90 min","Rated":"PG","Director":"Dave Green","Actors":"Will Forte, Lana Condor, John Cena","Plot":"A coyote buys a rocket.",` +
		`"Poster":"` + omdbAmazonPoster + `","Ratings":[{"Source":"Internet Movie Database","Value":"7.6/10"},{"Source":"Rotten Tomatoes","Value":"85%"},{"Source":"Metacritic","Value":"62/100"}]}`
)

// omdbLookupCase is one fake OMDb response, the poster width of the client, and the outcome that Lookup must produce for it.
type omdbLookupCase struct {
	name, body        string
	status            int
	delay             time.Duration
	posterWidth       int
	wantErr, notFound bool
	want              Metadata
}

func TestOMDbLookup(t *testing.T) {
	for _, tt := range []omdbLookupCase{
		{name: "found", body: omdbFoundBody, posterWidth: 1000, want: Metadata{Title: "Coyote vs. Acme", Year: "2021", IMDbID: "tt1756855", IMDbRating: "7.6", IMDbVotes: "12,345", RottenTomatoes: "85%", Metacritic: "62/100", Genre: "Animation", Runtime: "90 min", Rated: "PG", Director: "Dave Green", Actors: "Will Forte, Lana Condor, John Cena", Plot: "A coyote buys a rocket.", Poster: "https://m.media-amazon.com/images/M/MV5BNzk@._V1_SX1000.jpg", PosterFallback: omdbAmazonPoster}},
		{name: "missing", body: `{"Response":"True","Poster":"N/A","Plot":"N/A","Director":"N/A","imdbVotes":"N/A"}`},
		{name: "same width", body: `{"Response":"True","Poster":"` + omdbAmazonPoster + `"}`, posterWidth: 300, want: Metadata{Poster: omdbAmazonPoster}},
		{name: "unknown poster", body: `{"Response":"True","Poster":"https://example.com/poster.jpg"}`, posterWidth: 1000, want: Metadata{Poster: "https://example.com/poster.jpg"}},
		{name: "resizing off", body: `{"Response":"True","Poster":"` + omdbAmazonPoster + `"}`, want: Metadata{Poster: omdbAmazonPoster}},
		{name: "not found", body: `{"Response":"False"}`, wantErr: true, notFound: true},
		{name: "unauthorized", status: 401, wantErr: true}, {name: "invalid", body: `{`, wantErr: true}, {name: "timeout", delay: 100 * time.Millisecond, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := newOMDbTestServer(t, tt)
			defer s.Close()
			start := time.Now()
			m, err := NewOMDb(s.URL, "secret", 20*time.Millisecond, tt.posterWidth).Lookup(context.Background(), Release{Title: "Example", Year: "2021"})
			checkOMDbError(t, tt, err, time.Since(start))
			checkOMDbMetadata(t, tt, m)
		})
	}
}

// newOMDbTestServer serves the case's delay, status and body, and checks the lookup query parameters.
func newOMDbTestServer(t *testing.T, tt omdbLookupCase) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("t") != "Example" || r.URL.Query().Get("y") != "2021" || r.URL.Query().Get("apikey") != "secret" {
			t.Error("missing query parameters")
		}
		time.Sleep(tt.delay)
		if tt.status != 0 {
			w.WriteHeader(tt.status)
		}
		_, _ = fmt.Fprint(w, tt.body)
	}))
}

// checkOMDbError asserts the error outcome of one lookup, including the timeout bound and the key leak.
func checkOMDbError(t *testing.T, tt omdbLookupCase, err error, elapsed time.Duration) {
	t.Helper()
	if tt.name == "timeout" && elapsed >= tt.delay {
		t.Fatalf("lookup exceeded server delay: %v", elapsed)
	}
	if (err != nil) != tt.wantErr {
		t.Fatalf("error %v", err)
	}
	if err != nil && strings.Contains(err.Error(), "secret") {
		t.Fatal("key leaked")
	}
	if tt.notFound && !errors.Is(err, ErrNotFound) {
		t.Fatalf("unexpected error %v", err)
	}
}

// checkOMDbMetadata asserts that a successful lookup returns exactly the case's metadata.
func checkOMDbMetadata(t *testing.T, tt omdbLookupCase, m *Metadata) {
	t.Helper()
	if tt.wantErr {
		return
	}
	if !reflect.DeepEqual(*m, tt.want) {
		t.Fatalf("metadata %#v, want %#v", *m, tt.want)
	}
}
