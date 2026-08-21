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

const apiURL = "https://api.github.com"

// Client is the small subset of the GitHub API needed by automation-hub.
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
	logger     *zap.Logger
}

func NewClient(token string, logger *zap.Logger) *Client {
	return newClient(token, &http.Client{Timeout: 30 * time.Second}, apiURL, logger)
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
