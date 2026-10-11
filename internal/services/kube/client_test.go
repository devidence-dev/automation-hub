package kube

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

type failingRoundTripper struct{}

func (failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network unavailable")
}

func assertRestartDeploymentRequest(t *testing.T, r *http.Request) {
	t.Helper()

	if r.Method != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", r.Method)
	}
	if r.URL.Path != "/apis/apps/v1/namespaces/infrastructure/deployments/github-runner" {
		t.Errorf("path = %s", r.URL.Path)
	}
	if r.Header.Get("Authorization") != "Bearer test-token" {
		t.Errorf("Authorization header = %q", r.Header.Get("Authorization"))
	}
	if ct := r.Header.Get("Content-Type"); ct != "application/strategic-merge-patch+json" {
		t.Errorf("Content-Type = %q", ct)
	}

	body, _ := io.ReadAll(r.Body)
	var patch struct {
		Spec struct {
			Template struct {
				Metadata struct {
					Annotations map[string]string `json:"annotations"`
				} `json:"metadata"`
			} `json:"template"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(body, &patch); err != nil {
		t.Fatalf("invalid patch body: %v", err)
	}
	if _, ok := patch.Spec.Template.Metadata.Annotations["kubectl.kubernetes.io/restartedAt"]; !ok {
		t.Errorf("patch body missing restartedAt annotation: %s", body)
	}
}

func TestRestartDeployment(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{name: "success", statusCode: http.StatusOK},
		{name: "forbidden", statusCode: http.StatusForbidden, wantErr: true},
		{name: "not found", statusCode: http.StatusNotFound, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertRestartDeploymentRequest(t, r)
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
			err := client.RestartDeployment(context.Background(), "infrastructure", "github-runner")
			if (err != nil) != tt.wantErr {
				t.Fatalf("RestartDeployment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRestartDeploymentNetworkError(t *testing.T) {
	client := NewClientWithBaseURL("token", "https://example.test", &http.Client{Transport: failingRoundTripper{}}, zap.NewNop())
	err := client.RestartDeployment(context.Background(), "infrastructure", "github-runner")
	if err == nil || !strings.Contains(err.Error(), "network unavailable") {
		t.Fatalf("RestartDeployment() error = %v", err)
	}
}

func TestNewInClusterClientOutsideCluster(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("KUBERNETES_SERVICE_PORT", "")
	if _, err := NewInClusterClient(zap.NewNop()); err == nil {
		t.Fatal("NewInClusterClient() error = nil, want error when not running in a cluster")
	}
}

func deploymentJSON(generation, observedGeneration int64, replicas, updated, ready, available int32) string {
	body, _ := json.Marshal(map[string]any{
		"metadata": map[string]any{"generation": generation},
		"spec":     map[string]any{"replicas": replicas},
		"status": map[string]any{
			"observedGeneration": observedGeneration,
			"replicas":           replicas,
			"updatedReplicas":    updated,
			"readyReplicas":      ready,
			"availableReplicas":  available,
		},
	})
	return string(body)
}

func TestWaitForRolloutReturnsOnceReady(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/apis/apps/v1/namespaces/infrastructure/deployments/github-runner" {
			t.Errorf("path = %s", r.URL.Path)
		}
		n := atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusOK)
		if n < 3 {
			// Still rolling out: old pod terminating, new one not ready yet.
			_, _ = w.Write([]byte(deploymentJSON(2, 2, 1, 1, 0, 0)))
			return
		}
		_, _ = w.Write([]byte(deploymentJSON(2, 2, 1, 1, 1, 1)))
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	err := client.WaitForRollout(context.Background(), "infrastructure", "github-runner", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForRollout() error = %v", err)
	}
	if requests < 3 {
		t.Errorf("requests = %d, want at least 3 polls before ready", requests)
	}
}

func TestWaitForRolloutTimesOutWhileNotReady(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(deploymentJSON(2, 2, 1, 1, 0, 0)))
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := client.WaitForRollout(ctx, "infrastructure", "github-runner", 10*time.Millisecond)
	if err == nil {
		t.Fatal("WaitForRollout() error = nil, want a timeout error")
	}
}

func TestWaitForRolloutHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	err := client.WaitForRollout(context.Background(), "infrastructure", "github-runner", 10*time.Millisecond)
	if err == nil {
		t.Fatal("WaitForRollout() error = nil, want error")
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

func TestCreateJobFromCronJob(t *testing.T) {
	const cronJobBody = `{
	  "metadata": {"name": "dependency-report", "uid": "uid-123"},
	  "spec": {"suspend": true, "jobTemplate": {
	    "metadata": {"labels": {"app.kubernetes.io/name": "dependency-report"}, "annotations": {"team": "ops"}},
	    "spec": {"backoffLimit": 0, "futureField": {"x": 1}, "template": {"spec": {"restartPolicy": "Never"}}}
	  }}
	}`
	tests := []struct {
		name       string
		postStatus int
		wantErr    string
	}{
		{name: "created", postStatus: http.StatusCreated},
		{name: "forbidden", postStatus: http.StatusForbidden, wantErr: "403"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var posted map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
				}
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/apis/batch/v1/namespaces/internal-apps/cronjobs/dependency-report":
					_, _ = io.WriteString(w, cronJobBody)
				case r.Method == http.MethodPost && r.URL.Path == "/apis/batch/v1/namespaces/internal-apps/jobs":
					if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
						t.Errorf("decode job: %v", err)
					}
					w.WriteHeader(tt.postStatus)
					_, _ = io.WriteString(w, `{"message":"jobs is forbidden"}`)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
			}))
			defer server.Close()

			client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
			client.now = func() time.Time { return time.Unix(1791763200, 0) }
			name, err := client.CreateJobFromCronJob(context.Background(), "internal-apps", "dependency-report")

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateJobFromCronJob() error = %v", err)
			}
			if name != "dependency-report-manual-1791763200" {
				t.Errorf("name = %q", name)
			}
			meta := posted["metadata"].(map[string]any)
			if meta["name"] != name || meta["namespace"] != "internal-apps" {
				t.Errorf("metadata = %v", meta)
			}
			if got := meta["labels"].(map[string]any)["app.kubernetes.io/name"]; got != "dependency-report" {
				t.Errorf("label = %v", got)
			}
			ann := meta["annotations"].(map[string]any)
			if ann["cronjob.kubernetes.io/instantiate"] != "manual" || ann["team"] != "ops" {
				t.Errorf("annotations = %v", ann)
			}
			owner := meta["ownerReferences"].([]any)[0].(map[string]any)
			if owner["kind"] != "CronJob" || owner["name"] != "dependency-report" || owner["uid"] != "uid-123" || owner["controller"] != true {
				t.Errorf("ownerReference = %v", owner)
			}
			if _, set := owner["blockOwnerDeletion"]; set {
				t.Error("blockOwnerDeletion must not be set: kubectl omits it")
			}
			spec := posted["spec"].(map[string]any)
			if spec["backoffLimit"] != float64(0) || spec["futureField"] == nil {
				t.Errorf("job spec was not copied as-is: %v", spec)
			}
		})
	}
}

func TestCreateJobFromCronJobGetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	if _, err := client.CreateJobFromCronJob(context.Background(), "internal-apps", "missing"); err == nil || !strings.Contains(err.Error(), "get cronjob") {
		t.Fatalf("error = %v, want get cronjob failure", err)
	}
}
