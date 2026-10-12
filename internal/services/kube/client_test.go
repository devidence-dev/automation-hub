package kube

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"automation-hub/internal/httpx"

	"go.uber.org/zap"
)

type failingRoundTripper struct{}

func (failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network unavailable")
}

// fixedClock is 2026-10-11 09:07:00 UTC, which is 1791709620 in Unix seconds.
func fixedClock() time.Time {
	return time.Date(2026, time.October, 11, 9, 7, 0, 0, time.UTC)
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
	if ct := r.Header.Get("Content-Type"); ct != mediaTypeStrategicMergePatch {
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
	if got := patch.Spec.Template.Metadata.Annotations["kubectl.kubernetes.io/restartedAt"]; got != "2026-10-11T09:07:00Z" {
		t.Errorf("restartedAt = %q, want 2026-10-11T09:07:00Z (body %s)", got, body)
	}
}

func TestRestartDeployment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertRestartDeploymentRequest(t, r)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	client.now = fixedClock
	if err := client.RestartDeployment(context.Background(), "infrastructure", "github-runner"); err != nil {
		t.Fatalf("RestartDeployment() error = %v", err)
	}
}

func TestRestartDeploymentErrorStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{name: "forbidden", statusCode: http.StatusForbidden},
		{name: "not found", statusCode: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertRestartDeploymentRequest(t, r)
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
			client.now = fixedClock
			if err := client.RestartDeployment(context.Background(), "infrastructure", "github-runner"); err == nil {
				t.Fatal("RestartDeployment() error = nil, want error")
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
	if _, err := NewInClusterClient(zap.NewNop(), httpx.Settings{}, httpx.DefaultMaxResponseBytes); err == nil {
		t.Fatal("NewInClusterClient() error = nil, want error when not running in a cluster")
	}
}

func TestInClusterSettingsUseCA(t *testing.T) {
	server := httptest.NewTLSServer(http.NotFoundHandler())
	defer server.Close()
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})

	got, err := withServiceAccountCA(httpx.Settings{Timeout: 20 * time.Second, MaxIdleConns: 7}, caPEM)
	if err != nil {
		t.Fatalf("withServiceAccountCA() error = %v", err)
	}
	if got.Timeout != 20*time.Second || got.MaxIdleConns != 7 {
		t.Errorf("settings not kept: timeout %v, max idle %d", got.Timeout, got.MaxIdleConns)
	}
	if got.TLSConfig == nil || got.TLSConfig.RootCAs == nil {
		t.Fatal("TLSConfig.RootCAs not set")
	}
	opts := x509.VerifyOptions{Roots: got.TLSConfig.RootCAs, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	if _, err := server.Certificate().Verify(opts); err != nil {
		t.Errorf("CA pool does not trust the service account CA: %v", err)
	}

	if _, err := withServiceAccountCA(httpx.Settings{}, []byte("not a certificate")); err == nil {
		t.Error("withServiceAccountCA() accepted a non-PEM CA")
	}
}

func TestDoJSONCapsErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, strings.Repeat("x", 4096))
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	client.maxBytes = 1024
	err := client.doJSON(context.Background(), apiCall{method: http.MethodGet, endpoint: server.URL, wantStatus: http.StatusOK})
	if err == nil {
		t.Fatal("doJSON() error = nil, want status error")
	}
	if got := strings.Count(err.Error(), "x"); got != 1024 {
		t.Errorf("error carries %d body bytes, want 1024", got)
	}
}

// TestDoJSONAcceptsTrailingBytes checks that a 200 body is decoded from its first JSON value, as
// json.Decoder did before, so bytes after the object do not fail the call.
func TestDoJSONAcceptsTrailingBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"metadata":{"generation":7}} trailing bytes`)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	var out struct {
		Metadata struct {
			Generation int64 `json:"generation"`
		} `json:"metadata"`
	}
	err := client.doJSON(context.Background(), apiCall{method: http.MethodGet, endpoint: server.URL, wantStatus: http.StatusOK, out: &out})
	if err != nil {
		t.Fatalf("doJSON() error = %v, want the first JSON value decoded", err)
	}
	if out.Metadata.Generation != 7 {
		t.Errorf("generation = %d, want 7", out.Metadata.Generation)
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
	if client.maxBytes != httpx.DefaultMaxResponseBytes {
		t.Errorf("maxBytes = %d, want %d", client.maxBytes, httpx.DefaultMaxResponseBytes)
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
			runCreateJobCase(t, cronJobBody, tt.postStatus, tt.wantErr)
		})
	}
}

// runCreateJobCase runs CreateJobFromCronJob against a fake API server that
// serves body as the CronJob and answers the Job POST with postStatus.
func runCreateJobCase(t *testing.T, body string, postStatus int, wantErr string) {
	t.Helper()
	var posted map[string]any
	server := newCronJobAPIServer(t, body, postStatus, &posted)
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	client.now = func() time.Time { return time.Unix(1791763200, 0) }
	name, err := client.CreateJobFromCronJob(context.Background(), "internal-apps", "dependency-report")

	if wantErr != "" {
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("error = %v, want it to contain %q", err, wantErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("CreateJobFromCronJob() error = %v", err)
	}
	if name != "dependency-report-manual-1791763200" {
		t.Errorf("name = %q", name)
	}
	assertPostedJob(t, posted, name)
}

// newCronJobAPIServer serves the CronJob GET with body and answers the Job POST
// with postStatus, storing the posted Job in posted.
func newCronJobAPIServer(t *testing.T, body string, postStatus int, posted *map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/apis/batch/v1/namespaces/internal-apps/cronjobs/dependency-report":
			_, _ = io.WriteString(w, body)
		case r.Method == http.MethodPost && r.URL.Path == "/apis/batch/v1/namespaces/internal-apps/jobs":
			if err := json.NewDecoder(r.Body).Decode(posted); err != nil {
				t.Errorf("decode job: %v", err)
			}
			w.WriteHeader(postStatus)
			_, _ = io.WriteString(w, `{"message":"jobs is forbidden"}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
}

// assertPostedJob checks the Job posted for the dependency-report CronJob: its
// metadata, the copied template and the spec copied as-is.
func assertPostedJob(t *testing.T, posted map[string]any, name string) {
	t.Helper()
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

func TestCreateJobFromCronJobCapsCronJobBody(t *testing.T) {
	body := `{"metadata":{"name":"dependency-report","uid":"uid-123"},"spec":{"jobTemplate":{"spec":{"pad":"` +
		strings.Repeat("a", 4096) + `"}}}}`
	var posted map[string]any
	server := newCronJobAPIServer(t, body, http.StatusCreated, &posted)
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	client.maxBytes = 1024
	_, err := client.CreateJobFromCronJob(context.Background(), "internal-apps", "dependency-report")
	if err == nil || !strings.Contains(err.Error(), "decode response") {
		t.Fatalf("CreateJobFromCronJob() error = %v, want decode response failure", err)
	}
	if posted != nil {
		t.Errorf("a Job was posted for a truncated CronJob: %v", posted)
	}
}

func TestJobNameUsesInjectedClock(t *testing.T) {
	const cronJobBody = `{"metadata":{"name":"ops-cleanup","uid":"uid-456"},"spec":{"jobTemplate":{"spec":{}}}}`
	var posted map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/apis/batch/v1/namespaces/internal-apps/cronjobs/ops-cleanup":
			_, _ = io.WriteString(w, cronJobBody)
		case r.Method == http.MethodPost && r.URL.Path == "/apis/batch/v1/namespaces/internal-apps/jobs":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Errorf("decode job: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClientWithBaseURL("test-token", server.URL, server.Client(), zap.NewNop())
	client.now = fixedClock
	name, err := client.CreateJobFromCronJob(context.Background(), "internal-apps", "ops-cleanup")
	if err != nil {
		t.Fatalf("CreateJobFromCronJob() error = %v", err)
	}
	if name != "ops-cleanup-manual-1791709620" {
		t.Errorf("name = %q, want ops-cleanup-manual-1791709620", name)
	}
	if meta, _ := posted["metadata"].(map[string]any); meta["name"] != name {
		t.Errorf("posted metadata = %v", posted["metadata"])
	}
}
