package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"automation-hub/internal/httpx"

	"go.uber.org/zap"
)

// defaultTestTimeout is the request timeout of the NewClientWithBaseURL fallback client, the 30s
// of the default transport this client used before. Production clients take github.http.timeout.
const defaultTestTimeout = 30 * time.Second

// Config holds the settings of the GitHub client. The caller fills it from its own configuration.
type Config struct {
	Token            string
	APIURL           string // base of the REST API; a trailing slash is trimmed
	HTTP             httpx.Settings
	MaxResponseBytes int64 // cap on every response body the client reads
}

// Client is the small subset of the GitHub API needed by automation-hub.
type Client struct {
	token            string
	httpClient       *http.Client
	baseURL          string
	maxResponseBytes int64
	logger           *zap.Logger
}

// NewClient builds a client whose transport comes from cfg.HTTP and whose requests go to cfg.APIURL.
func NewClient(cfg Config, logger *zap.Logger) *Client {
	return newClient(cfg.Token, httpx.NewClient(cfg.HTTP), cfg.APIURL, cfg.MaxResponseBytes, logger)
}

// NewClientWithBaseURL lets tests point the client at an httptest.Server.
func NewClientWithBaseURL(token, baseURL string, httpClient *http.Client, logger *zap.Logger) *Client {
	if httpClient == nil {
		httpClient = httpx.NewClient(httpx.Settings{Timeout: defaultTestTimeout})
	}
	return newClient(token, httpClient, baseURL, httpx.DefaultMaxResponseBytes, logger)
}

func newClient(token string, httpClient *http.Client, baseURL string, maxResponseBytes int64, logger *zap.Logger) *Client {
	if logger == nil {
		logger = zap.NewNop()
	}
	// A zero cap would silently empty every body, so fall back to the shared default.
	if maxResponseBytes <= 0 {
		maxResponseBytes = httpx.DefaultMaxResponseBytes
	}
	return &Client{
		token:            token,
		httpClient:       httpClient,
		baseURL:          strings.TrimRight(baseURL, "/"),
		maxResponseBytes: maxResponseBytes,
		logger:           logger,
	}
}

// DispatchWorkflow starts a workflow_dispatch run and returns the HTML URL of
// the run that GitHub created. GitHub answers 204 without run details when
// the request does not get them back, so the URL is empty with no error in
// that case, and the caller falls back to the repository's Actions page.
func (c *Client) DispatchWorkflow(ctx context.Context, owner, repo, workflowFile, ref string) (string, error) {
	req, err := c.newDispatchRequest(owner, repo, workflowFile, ref)
	if err != nil {
		return "", err
	}
	resp, err := httpx.Send(ctx, c.httpClient, req, c.maxResponseBytes)
	if err != nil {
		return "", fmt.Errorf("dispatch workflow request: %w", err)
	}
	return c.decodeDispatch(resp)
}

// newDispatchRequest builds the POST that asks GitHub to start the workflow.
func (c *Client) newDispatchRequest(owner, repo, workflowFile, ref string) (httpx.Request, error) {
	body, err := json.Marshal(struct {
		Ref              string `json:"ref"`
		ReturnRunDetails bool   `json:"return_run_details"`
	}{Ref: ref, ReturnRunDetails: true})
	if err != nil {
		return httpx.Request{}, fmt.Errorf("marshal workflow dispatch request: %w", err)
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.token)
	header.Set("Accept", "application/vnd.github+json")
	header.Set("Content-Type", "application/json")

	return httpx.Request{
		Method: http.MethodPost,
		URL: fmt.Sprintf("%s/repos/%s/%s/actions/workflows/%s/dispatches",
			c.baseURL, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(workflowFile)),
		Header: header,
		Body:   body,
	}, nil
}

// decodeDispatch maps the dispatch answer to the run URL: 204 has none, 200 carries it,
// and any other status is an error.
func (c *Client) decodeDispatch(resp httpx.Response) (string, error) {
	switch resp.StatusCode {
	case http.StatusNoContent:
		return "", nil
	case http.StatusOK:
		var parsed struct {
			HTMLURL string `json:"html_url"`
		}
		// The run already exists at this point: report success without a URL
		// rather than an error that would invite a duplicate dispatch.
		if err := json.NewDecoder(bytes.NewReader(resp.Body)).Decode(&parsed); err != nil {
			c.logger.Warn("Workflow dispatched but run details could not be decoded", zap.Error(err))
			return "", nil
		}
		return parsed.HTMLURL, nil
	}
	return "", fmt.Errorf("GitHub workflow dispatch failed: status %s: %s", resp.Status, strings.TrimSpace(string(resp.Body)))
}
