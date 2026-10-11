// Package kube talks to the Kubernetes API server from inside a pod, using
// the projected ServiceAccount token instead of a full client-go dependency —
// automation-hub only ever needs a single PATCH endpoint.
package kube

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

const serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"

// Client is the small subset of the Kubernetes API needed by automation-hub.
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
	logger     *zap.Logger
	now        func() time.Time // overridden in tests
}

// NewInClusterClient builds a Client from the ServiceAccount token, CA
// certificate, and API server address that Kubernetes projects into every pod.
func NewInClusterClient(logger *zap.Logger) (*Client, error) {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return nil, fmt.Errorf("not running inside a Kubernetes pod: KUBERNETES_SERVICE_HOST/PORT not set")
	}

	token, err := os.ReadFile(serviceAccountDir + "/token")
	if err != nil {
		return nil, fmt.Errorf("read service account token: %w", err)
	}

	caCert, err := os.ReadFile(serviceAccountDir + "/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("read service account CA certificate: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("parse service account CA certificate")
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool},
		},
	}

	return newClient(strings.TrimSpace(string(token)), httpClient, "https://"+net.JoinHostPort(host, port), logger), nil
}

// NewClientWithBaseURL lets tests point the client at an httptest.Server.
func NewClientWithBaseURL(token, baseURL string, httpClient *http.Client, logger *zap.Logger) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return newClient(token, httpClient, baseURL, logger)
}

func newClient(token string, httpClient *http.Client, baseURL string, logger *zap.Logger) *Client {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Client{token: token, httpClient: httpClient, baseURL: strings.TrimRight(baseURL, "/"), logger: logger, now: time.Now}
}

// RestartDeployment triggers a rollout restart, the same effect as
// `kubectl rollout restart deployment/<name> -n <namespace>`: it patches the
// pod template with a restartedAt annotation, which the Deployment controller
// treats as a template change and rolls new pods out to satisfy.
func (c *Client) RestartDeployment(ctx context.Context, namespace, name string) error {
	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`,
		time.Now().UTC().Format(time.RFC3339))

	endpoint := fmt.Sprintf("%s/apis/apps/v1/namespaces/%s/deployments/%s",
		c.baseURL, url.PathEscape(namespace), url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, strings.NewReader(patch))
	if err != nil {
		return fmt.Errorf("create restart deployment request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/strategic-merge-patch+json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("restart deployment request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("restart deployment failed: status %s", resp.Status)
}

type deploymentStatus struct {
	Metadata struct {
		Generation int64 `json:"generation"`
	} `json:"metadata"`
	Spec struct {
		Replicas *int32 `json:"replicas"`
	} `json:"spec"`
	Status struct {
		ObservedGeneration int64 `json:"observedGeneration"`
		Replicas           int32 `json:"replicas"`
		UpdatedReplicas    int32 `json:"updatedReplicas"`
		ReadyReplicas      int32 `json:"readyReplicas"`
		AvailableReplicas  int32 `json:"availableReplicas"`
	} `json:"status"`
}

func (c *Client) getDeployment(ctx context.Context, namespace, name string) (deploymentStatus, error) {
	endpoint := fmt.Sprintf("%s/apis/apps/v1/namespaces/%s/deployments/%s",
		c.baseURL, url.PathEscape(namespace), url.PathEscape(name))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return deploymentStatus{}, fmt.Errorf("create get deployment request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return deploymentStatus{}, fmt.Errorf("get deployment request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return deploymentStatus{}, fmt.Errorf("get deployment failed: status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var parsed deploymentStatus
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return deploymentStatus{}, fmt.Errorf("decode deployment response: %w", err)
	}
	return parsed, nil
}

func (d deploymentStatus) rolloutComplete() bool {
	wantReplicas := int32(1)
	if d.Spec.Replicas != nil {
		wantReplicas = *d.Spec.Replicas
	}
	return d.Status.ObservedGeneration >= d.Metadata.Generation &&
		d.Status.UpdatedReplicas == wantReplicas &&
		d.Status.ReadyReplicas == wantReplicas &&
		d.Status.AvailableReplicas == wantReplicas
}

// WaitForRollout blocks until the Deployment's rollout finishes — the same
// condition `kubectl rollout status` waits for: the controller has observed
// the latest template generation, and updated/ready/available replicas all
// match the desired replica count. It returns ctx.Err() if ctx is cancelled
// or its deadline passes before that happens.
func (c *Client) WaitForRollout(ctx context.Context, namespace, name string, pollInterval time.Duration) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		status, err := c.getDeployment(ctx, namespace, name)
		if err != nil {
			return err
		}
		if status.rolloutComplete() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// cronJob is the part of a CronJob that CreateJobFromCronJob copies. The Job
// spec stays raw so fields this client does not know are not dropped.
type cronJob struct {
	Metadata struct {
		Name string `json:"name"`
		UID  string `json:"uid"`
	} `json:"metadata"`
	Spec struct {
		JobTemplate struct {
			Metadata struct {
				Labels      map[string]string `json:"labels"`
				Annotations map[string]string `json:"annotations"`
			} `json:"metadata"`
			Spec json.RawMessage `json:"spec"`
		} `json:"jobTemplate"`
	} `json:"spec"`
}

// CreateJobFromCronJob creates a Job from the CronJob's job template and
// returns its name. It mirrors `kubectl create job --from=cronjob/<name>`
// (createJobFromCronJob in kubectl's pkg/cmd/create/create_job.go): same
// labels, annotations plus cronjob.kubernetes.io/instantiate=manual, the
// template spec, and a controller ownerReference to the CronJob without
// blockOwnerDeletion. The name is <cronjob>-manual-<unix seconds>.
func (c *Client) CreateJobFromCronJob(ctx context.Context, namespace, name string) (string, error) {
	cronJobURL := fmt.Sprintf("%s/apis/batch/v1/namespaces/%s/cronjobs/%s",
		c.baseURL, url.PathEscape(namespace), url.PathEscape(name))
	var cj cronJob
	if err := c.doJSON(ctx, http.MethodGet, cronJobURL, nil, http.StatusOK, &cj); err != nil {
		return "", fmt.Errorf("get cronjob: %w", err)
	}

	annotations := map[string]string{"cronjob.kubernetes.io/instantiate": "manual"}
	for k, v := range cj.Spec.JobTemplate.Metadata.Annotations {
		annotations[k] = v
	}
	jobName := fmt.Sprintf("%s-manual-%d", name, c.now().Unix())
	job := map[string]any{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata": map[string]any{
			"name":        jobName,
			"namespace":   namespace,
			"labels":      cj.Spec.JobTemplate.Metadata.Labels,
			"annotations": annotations,
			"ownerReferences": []map[string]any{{
				"apiVersion": "batch/v1",
				"kind":       "CronJob",
				"name":       cj.Metadata.Name,
				"uid":        cj.Metadata.UID,
				"controller": true,
			}},
		},
		"spec": cj.Spec.JobTemplate.Spec,
	}
	body, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("encode job: %w", err)
	}

	jobsURL := fmt.Sprintf("%s/apis/batch/v1/namespaces/%s/jobs", c.baseURL, url.PathEscape(namespace))
	if err := c.doJSON(ctx, http.MethodPost, jobsURL, body, http.StatusCreated, nil); err != nil {
		return "", fmt.Errorf("create job: %w", err)
	}
	return jobName, nil
}

// doJSON sends a request and decodes the JSON response into out when out is
// not nil. Any status other than wantStatus is an error that carries the body.
func (c *Client) doJSON(ctx context.Context, method, endpoint string, body []byte, wantStatus int, out any) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != wantStatus {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("status %s: %s", resp.Status, strings.TrimSpace(string(msg)))
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
