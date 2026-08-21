package github

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

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
