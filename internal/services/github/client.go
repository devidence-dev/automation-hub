package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

var githubAPIURL = "https://api.github.com"

// Client is the small subset of the GitHub API needed by automation-hub.
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
	logger     *zap.Logger
}

func NewClient(token string, logger *zap.Logger) *Client {
	return newClient(token, &http.Client{Timeout: 30 * time.Second}, githubAPIURL, logger)
}

// NewClientWithBaseURL lets tests point the client at an httptest.Server.
func NewClientWithBaseURL(token, baseURL string, httpClient *http.Client, logger *zap.Logger) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return newClient(token, httpClient, baseURL, logger)
}

func newClient(token string, httpClient *http.Client, baseURL string, logger *zap.Logger) *Client {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Client{token: token, httpClient: httpClient, baseURL: strings.TrimRight(baseURL, "/"), logger: logger}
}

func (c *Client) DispatchWorkflow(ctx context.Context, owner, repo, workflowFile, ref string) error {
	body, err := json.Marshal(struct {
		Ref string `json:"ref"`
	}{Ref: ref})
	if err != nil {
		return fmt.Errorf("marshal workflow dispatch request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/repos/%s/%s/actions/workflows/%s/dispatches",
		c.baseURL, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(workflowFile))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create workflow dispatch request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dispatch workflow request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	responseBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if readErr != nil {
		return fmt.Errorf("read workflow dispatch response: %w", readErr)
	}
	return fmt.Errorf("GitHub workflow dispatch failed: status %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
}

type workflowRun struct {
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
}

type workflowRunsResponse struct {
	WorkflowRuns []workflowRun `json:"workflow_runs"`
}

// FindLatestRunURL looks up the most recent workflow_dispatch run created at
// or after "after" and returns its HTML URL. workflow_dispatch itself does
// not return the run it created, so callers poll this shortly afterwards.
// It returns an empty string (no error) if no matching run is found yet.
func (c *Client) FindLatestRunURL(ctx context.Context, owner, repo, workflowFile string, after time.Time) (string, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/actions/workflows/%s/runs?event=workflow_dispatch&per_page=5",
		c.baseURL, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(workflowFile))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create list workflow runs request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("list workflow runs request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return "", fmt.Errorf("list workflow runs failed: status %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}

	var parsed workflowRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode workflow runs response: %w", err)
	}

	for _, run := range parsed.WorkflowRuns {
		if !run.CreatedAt.Before(after) {
			return run.HTMLURL, nil
		}
	}
	return "", nil
}
