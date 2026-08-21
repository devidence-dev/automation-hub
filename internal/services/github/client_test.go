package github

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

type failingRoundTripper struct{}

func (failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network unavailable")
}

func TestDispatchWorkflow(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
	}{
		{name: "success", statusCode: http.StatusNoContent},
		{name: "unauthorized", statusCode: http.StatusUnauthorized, response: `{"message":"Bad credentials"}`, wantErr: true},
		{name: "not found", statusCode: http.StatusNotFound, response: `{"message":"Not Found"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != "/repos/owner/repo/actions/workflows/check-updates.yml/dispatches" {
					t.Errorf("path = %s", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("Authorization header = %q", r.Header.Get("Authorization"))
				}
				body, _ := io.ReadAll(r.Body)
				if string(body) != `{"ref":"master"}` {
					t.Errorf("body = %s", body)
				}
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
			err := client.DispatchWorkflow(context.Background(), "owner", "repo", "check-updates.yml", "master")
			if (err != nil) != tt.wantErr {
				t.Fatalf("DispatchWorkflow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewClientDispatchWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	previousAPIURL := githubAPIURL
	githubAPIURL = server.URL
	t.Cleanup(func() { githubAPIURL = previousAPIURL })

	client := NewClient("test-token", zap.NewNop())
	if err := client.DispatchWorkflow(context.Background(), "owner", "repo", "workflow.yml", "main"); err != nil {
		t.Fatalf("DispatchWorkflow() error = %v", err)
	}
}

func TestNewClientWithBaseURLUsesDefaults(t *testing.T) {
	client := NewClientWithBaseURL("token", "https://example.test/", nil, nil)
	if client.httpClient == nil || client.logger == nil {
		t.Fatal("NewClientWithBaseURL() did not initialize defaults")
	}
	if client.baseURL != "https://example.test" {
		t.Errorf("baseURL = %q", client.baseURL)
	}
}

func TestDispatchWorkflowNetworkError(t *testing.T) {
	client := NewClientWithBaseURL("token", "https://example.test", &http.Client{Transport: failingRoundTripper{}}, zap.NewNop())
	err := client.DispatchWorkflow(context.Background(), "owner", "repo", "workflow.yml", "main")
	if err == nil || !strings.Contains(err.Error(), "network unavailable") {
		t.Fatalf("DispatchWorkflow() error = %v", err)
	}
}

func TestFindLatestRunURL(t *testing.T) {
	after := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		response   string
		wantRunURL string
		wantErr    bool
	}{
		{
			name: "returns the most recent run created at or after the cutoff",
			response: `{"workflow_runs":[
				{"html_url":"https://github.com/owner/repo/actions/runs/2","created_at":"2026-08-21T12:00:05Z"},
				{"html_url":"https://github.com/owner/repo/actions/runs/1","created_at":"2026-08-21T11:59:00Z"}
			]}`,
			wantRunURL: "https://github.com/owner/repo/actions/runs/2",
		},
		{
			name:       "returns empty string when no run is new enough yet",
			response:   `{"workflow_runs":[{"html_url":"https://github.com/owner/repo/actions/runs/1","created_at":"2026-08-21T11:59:00Z"}]}`,
			wantRunURL: "",
		},
		{
			name:       "returns empty string when there are no runs yet",
			response:   `{"workflow_runs":[]}`,
			wantRunURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %s, want GET", r.Method)
				}
				if r.URL.Path != "/repos/owner/repo/actions/workflows/check-updates.yml/runs" {
					t.Errorf("path = %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
			runURL, err := client.FindLatestRunURL(context.Background(), "owner", "repo", "check-updates.yml", after)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FindLatestRunURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if runURL != tt.wantRunURL {
				t.Errorf("FindLatestRunURL() = %q, want %q", runURL, tt.wantRunURL)
			}
		})
	}
}

func TestFindLatestRunURLHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	_, err := client.FindLatestRunURL(context.Background(), "owner", "repo", "check-updates.yml", time.Now())
	if err == nil {
		t.Fatal("FindLatestRunURL() error = nil, want error")
	}
}
