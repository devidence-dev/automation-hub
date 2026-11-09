package media

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// omdbLookupCase is one fake OMDb response and the outcome that Lookup must produce for it.
type omdbLookupCase struct {
	name, body        string
	status            int
	delay             time.Duration
	wantErr, notFound bool
}

func TestOMDbLookup(t *testing.T) {
	for _, tt := range []omdbLookupCase{
		{name: "found", body: `{"Response":"True","Title":"Example","imdbID":"tt123","imdbRating":"7.3","Ratings":[{"Source":"Rotten Tomatoes","Value":"85%"}]}`},
		{name: "missing", body: `{"Response":"True","Poster":"N/A"}`},
		{name: "not found", body: `{"Response":"False"}`, wantErr: true, notFound: true},
		{name: "unauthorized", status: 401, wantErr: true}, {name: "invalid", body: `{`, wantErr: true}, {name: "timeout", delay: 100 * time.Millisecond, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := newOMDbTestServer(t, tt)
			defer s.Close()
			start := time.Now()
			m, err := NewOMDb(s.URL, "secret", 20*time.Millisecond).Lookup(context.Background(), Release{Title: "Example", Year: "2021"})
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

// checkOMDbMetadata asserts the fields that the found and missing cases must or must not carry.
func checkOMDbMetadata(t *testing.T, tt omdbLookupCase, m *Metadata) {
	t.Helper()
	if tt.name == "found" && (m.IMDbID != "tt123" || m.RottenTomatoes != "85%") {
		t.Fatalf("metadata %#v", m)
	}
	if tt.name == "missing" && (m.Poster != "" || m.RottenTomatoes != "") {
		t.Fatalf("metadata %#v", m)
	}
}
