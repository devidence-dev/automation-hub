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

	"automation-hub/internal/httpx"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type failingRoundTripper struct{}

func (failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network unavailable")
}

func assertDispatchWorkflowRequest(t *testing.T, r *http.Request) {
	t.Helper()

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
	if string(body) != `{"ref":"master","return_run_details":true}` {
		t.Errorf("body = %s", body)
	}
}

func TestDispatchWorkflow(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantURL    string
		wantErr    bool
	}{
		{name: "no content has no run details", statusCode: http.StatusNoContent},
		{
			name:       "ok returns the run html_url",
			statusCode: http.StatusOK,
			response:   `{"workflow_run_id":123456,"run_url":"https://api.github.com/repos/owner/repo/actions/runs/123456","html_url":"https://github.com/owner/repo/actions/runs/123456"}`,
			wantURL:    "https://github.com/owner/repo/actions/runs/123456",
		},
		{name: "ok with invalid json still counts as dispatched", statusCode: http.StatusOK, response: `not json`},
		{name: "unauthorized", statusCode: http.StatusUnauthorized, response: `{"message":"Bad credentials"}`, wantErr: true},
		{name: "not found", statusCode: http.StatusNotFound, response: `{"message":"Not Found"}`, wantErr: true},
		{name: "unprocessable entity", statusCode: http.StatusUnprocessableEntity, response: `{"message":"Ref does not exist"}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertDispatchWorkflowRequest(t, r)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
			runURL, err := client.DispatchWorkflow(context.Background(), "owner", "repo", "check-updates.yml", "master")
			if (err != nil) != tt.wantErr {
				t.Fatalf("DispatchWorkflow() error = %v, wantErr %v", err, tt.wantErr)
			}
			if runURL != tt.wantURL {
				t.Errorf("DispatchWorkflow() url = %q, want %q", runURL, tt.wantURL)
			}
		})
	}
}

func TestNewClientDispatchWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(Config{Token: "test-token", APIURL: server.URL, MaxResponseBytes: httpx.DefaultMaxResponseBytes}, zap.NewNop())
	if _, err := client.DispatchWorkflow(context.Background(), "owner", "repo", "workflow.yml", "main"); err != nil {
		t.Fatalf("DispatchWorkflow() error = %v", err)
	}
}

func TestNewClientUsesAPIURL(t *testing.T) {
	const wantPath = "/api/v3/repos/owner/repo/actions/workflows/check-updates.yml/dispatches"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("path = %s, want %s", r.URL.Path, wantPath)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(Config{Token: "test-token", APIURL: server.URL + "/api/v3/", MaxResponseBytes: httpx.DefaultMaxResponseBytes}, zap.NewNop())
	if _, err := client.DispatchWorkflow(context.Background(), "owner", "repo", "check-updates.yml", "master"); err != nil {
		t.Fatalf("DispatchWorkflow() error = %v", err)
	}
}

func TestDispatchWorkflowCapsOversizedBody(t *testing.T) {
	const capBytes = 1024
	// The run URL comes first, so only a decoder that sees the whole body could fail on the cut.
	head := `{"html_url":"https://github.com/owner/repo/actions/runs/1","padding":"`
	tail := `"}`
	body := head + strings.Repeat("x", 2000-len(head)-len(tail)) + tail
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	core, logs := observer.New(zap.WarnLevel)
	client := NewClient(Config{Token: "test-token", APIURL: server.URL, MaxResponseBytes: capBytes}, zap.New(core))
	runURL, err := client.DispatchWorkflow(context.Background(), "owner", "repo", "check-updates.yml", "master")
	if err != nil {
		t.Fatalf("DispatchWorkflow() error = %v", err)
	}
	if runURL != "" {
		t.Errorf("DispatchWorkflow() url = %q, want empty for a truncated body", runURL)
	}
	if got := logs.FilterMessage("Workflow dispatched but run details could not be decoded").Len(); got != 1 {
		t.Errorf("undecodable-body warnings = %d, want 1", got)
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
	if client.maxResponseBytes != httpx.DefaultMaxResponseBytes {
		t.Errorf("maxResponseBytes = %d, want %d", client.maxResponseBytes, httpx.DefaultMaxResponseBytes)
	}
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("timeout = %s, want 30s", client.httpClient.Timeout)
	}
}

func TestDispatchWorkflowNetworkError(t *testing.T) {
	client := NewClientWithBaseURL("token", "https://example.test", &http.Client{Transport: failingRoundTripper{}}, zap.NewNop())
	_, err := client.DispatchWorkflow(context.Background(), "owner", "repo", "workflow.yml", "main")
	if err == nil || !strings.Contains(err.Error(), "network unavailable") {
		t.Fatalf("DispatchWorkflow() error = %v", err)
	}
}
