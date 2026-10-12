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
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"automation-hub/internal/httpx"

	"go.uber.org/zap"
)

const serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"

const (
	bearerPrefix                 = "Bearer "
	mediaTypeJSON                = "application/json"
	mediaTypeStrategicMergePatch = "application/strategic-merge-patch+json"
	batchAPIVersion              = "batch/v1"
)

// API paths relative to the API server. resourceURL escapes the namespace and name.
const (
	deploymentPath = "/apis/apps/v1/namespaces/%s/deployments/%s"
	cronJobPath    = "/apis/batch/v1/namespaces/%s/cronjobs/%s"
	jobsPath       = "/apis/batch/v1/namespaces/%s/jobs"
)

// jobNameFormat is <cronjob>-manual-<unix seconds>.
const jobNameFormat = "%s-manual-%d"

// defaultTestTimeout is the request timeout of the NewClientWithBaseURL fallback client.
// Production clients take their timeout from kubernetes.http.timeout.
const defaultTestTimeout = 15 * time.Second

// Client is the small subset of the Kubernetes API needed by automation-hub.
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
	maxBytes   int64 // caps every response body the client reads
	logger     *zap.Logger
	now        func() time.Time // overridden in tests
}

// NewInClusterClient builds a Client from the ServiceAccount token, CA
// certificate, and API server address that Kubernetes projects into every pod.
// settings are the transport values of kubernetes.http; the TLS config is always
// the ServiceAccount CA. maxResponseBytes caps every response body.
func NewInClusterClient(logger *zap.Logger, settings httpx.Settings, maxResponseBytes int64) (*Client, error) {
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
	settings, err = withServiceAccountCA(settings, caCert)
	if err != nil {
		return nil, err
	}

	httpClient := httpx.NewClient(settings)
	return newClient(strings.TrimSpace(string(token)), httpClient, "https://"+net.JoinHostPort(host, port), maxResponseBytes, logger), nil
}

// withServiceAccountCA returns settings whose TLS config trusts only the certificates in caPEM.
func withServiceAccountCA(settings httpx.Settings, caPEM []byte) (httpx.Settings, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return httpx.Settings{}, fmt.Errorf("parse service account CA certificate")
	}
	settings.TLSConfig = &tls.Config{RootCAs: pool}
	return settings, nil
}

// NewClientWithBaseURL lets tests point the client at an httptest.Server.
func NewClientWithBaseURL(token, baseURL string, httpClient *http.Client, logger *zap.Logger) *Client {
	if httpClient == nil {
		httpClient = httpx.NewClient(httpx.Settings{Timeout: defaultTestTimeout})
	}
	return newClient(token, httpClient, baseURL, httpx.DefaultMaxResponseBytes, logger)
}

func newClient(token string, httpClient *http.Client, baseURL string, maxBytes int64, logger *zap.Logger) *Client {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Client{token: token, httpClient: httpClient, baseURL: strings.TrimRight(baseURL, "/"), maxBytes: maxBytes, logger: logger, now: time.Now}
}

// RestartDeployment triggers a rollout restart, the same effect as
// `kubectl rollout restart deployment/<name> -n <namespace>`: it patches the
// pod template with a restartedAt annotation, which the Deployment controller
// treats as a template change and rolls new pods out to satisfy.
func (c *Client) RestartDeployment(ctx context.Context, namespace, name string) error {
	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`,
		c.now().UTC().Format(time.RFC3339))

	err := c.doJSON(ctx, apiCall{
		method:      http.MethodPatch,
		endpoint:    resourceURL(c.baseURL, deploymentPath, namespace, name),
		contentType: mediaTypeStrategicMergePatch,
		body:        []byte(patch),
		wantStatus:  http.StatusOK,
	})
	if err != nil {
		return fmt.Errorf("restart deployment: %w", err)
	}
	return nil
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
	var parsed deploymentStatus
	err := c.doJSON(ctx, apiCall{
		method:     http.MethodGet,
		endpoint:   resourceURL(c.baseURL, deploymentPath, namespace, name),
		wantStatus: http.StatusOK,
		out:        &parsed,
	})
	if err != nil {
		return deploymentStatus{}, fmt.Errorf("get deployment: %w", err)
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
		JobTemplate cronJobTemplate `json:"jobTemplate"`
	} `json:"spec"`
}

// cronJobTemplate is the job template of a CronJob spec.
type cronJobTemplate struct {
	Metadata cronJobTemplateMetadata `json:"metadata"`
	Spec     json.RawMessage         `json:"spec"`
}

// cronJobTemplateMetadata holds the labels and annotations copied to the Job.
type cronJobTemplateMetadata struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// job is the batch/v1 Job that CreateJobFromCronJob posts. Spec stays raw, as in cronJob.
type job struct {
	APIVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Metadata   objectMeta      `json:"metadata"`
	Spec       json.RawMessage `json:"spec"`
}

// objectMeta is the metadata of the posted Job.
type objectMeta struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	OwnerReferences []ownerReference  `json:"ownerReferences"`
}

// ownerReference points the Job at the CronJob that created it.
type ownerReference struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	UID        string `json:"uid"`
	Controller bool   `json:"controller"`
}

// CreateJobFromCronJob creates a Job from the CronJob's job template and
// returns its name. It mirrors `kubectl create job --from=cronjob/<name>`
// (createJobFromCronJob in kubectl's pkg/cmd/create/create_job.go): same
// labels, annotations plus cronjob.kubernetes.io/instantiate=manual, the
// template spec, and a controller ownerReference to the CronJob without
// blockOwnerDeletion. The name is <cronjob>-manual-<unix seconds>.
func (c *Client) CreateJobFromCronJob(ctx context.Context, namespace, name string) (string, error) {
	var cj cronJob
	err := c.doJSON(ctx, apiCall{
		method:     http.MethodGet,
		endpoint:   resourceURL(c.baseURL, cronJobPath, namespace, name),
		wantStatus: http.StatusOK,
		out:        &cj,
	})
	if err != nil {
		return "", fmt.Errorf("get cronjob: %w", err)
	}

	jobName := fmt.Sprintf(jobNameFormat, name, c.now().Unix())
	body, err := json.Marshal(jobFromTemplate(cj, namespace, jobName))
	if err != nil {
		return "", fmt.Errorf("encode job: %w", err)
	}

	err = c.doJSON(ctx, apiCall{
		method:      http.MethodPost,
		endpoint:    c.baseURL + fmt.Sprintf(jobsPath, url.PathEscape(namespace)),
		contentType: mediaTypeJSON,
		body:        body,
		wantStatus:  http.StatusCreated,
	})
	if err != nil {
		return "", fmt.Errorf("create job: %w", err)
	}
	return jobName, nil
}

// jobFromTemplate builds the Job named name from the CronJob's job template.
func jobFromTemplate(cj cronJob, namespace, name string) job {
	annotations := map[string]string{"cronjob.kubernetes.io/instantiate": "manual"}
	for k, v := range cj.Spec.JobTemplate.Metadata.Annotations {
		annotations[k] = v
	}
	return job{
		APIVersion: batchAPIVersion,
		Kind:       "Job",
		Metadata: objectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      cj.Spec.JobTemplate.Metadata.Labels,
			Annotations: annotations,
			OwnerReferences: []ownerReference{{
				APIVersion: batchAPIVersion,
				Kind:       "CronJob",
				Name:       cj.Metadata.Name,
				UID:        cj.Metadata.UID,
				Controller: true,
			}},
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}
}

// apiCall describes one request for doJSON. contentType and body are empty for a GET.
// A nil out means the response body is not decoded.
type apiCall struct {
	method      string
	endpoint    string
	contentType string
	body        []byte
	wantStatus  int
	out         any
}

// doJSON sends the call through httpx.Send, with the response body capped at c.maxBytes.
// A status other than wantStatus is an error that carries the capped body. The response is
// decoded into call.out when it is not nil.
func (c *Client) doJSON(ctx context.Context, call apiCall) error {
	header := http.Header{}
	header.Set("Authorization", bearerPrefix+c.token)
	header.Set("Accept", mediaTypeJSON)
	if call.contentType != "" {
		header.Set("Content-Type", call.contentType)
	}

	resp, err := httpx.Send(ctx, c.httpClient, httpx.Request{
		Method: call.method,
		URL:    call.endpoint,
		Header: header,
		Body:   call.body,
	}, c.maxBytes)
	if err != nil {
		return err
	}

	if resp.StatusCode != call.wantStatus {
		return fmt.Errorf("status %s: %s", resp.Status, strings.TrimSpace(string(resp.Body)))
	}
	if call.out == nil {
		return nil
	}
	if err := json.NewDecoder(bytes.NewReader(resp.Body)).Decode(call.out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// resourceURL fills a path format with the namespace and name, escaping each as a path segment.
func resourceURL(baseURL, pathFormat, namespace, name string) string {
	return baseURL + fmt.Sprintf(pathFormat, url.PathEscape(namespace), url.PathEscape(name))
}
