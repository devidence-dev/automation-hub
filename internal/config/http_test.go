package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"automation-hub/internal/httpx"
)

func TestServerTimeoutDefaults(t *testing.T) {
	tests := []struct {
		name                            string
		content                         string
		wantRead, wantWrite, wantShutdn time.Duration
	}{
		{name: "defaults when omitted", content: serverOnly, wantRead: 15 * time.Second, wantWrite: 15 * time.Second, wantShutdn: 30 * time.Second},
		{
			name:       "read timeout raised",
			content:    "server:\n  address: \":8080\"\n  read_timeout: \"30s\"\n",
			wantRead:   30 * time.Second,
			wantWrite:  15 * time.Second,
			wantShutdn: 30 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loadYAML(t, tt.content)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.Server.ReadTimeout != tt.wantRead || cfg.Server.WriteTimeout != tt.wantWrite || cfg.Server.ShutdownTimeout != tt.wantShutdn {
				t.Errorf("server timeouts = %s, %s, %s; want %s, %s, %s", cfg.Server.ReadTimeout, cfg.Server.WriteTimeout,
					cfg.Server.ShutdownTimeout, tt.wantRead, tt.wantWrite, tt.wantShutdn)
			}
		})
	}
}

func TestServerTimeoutRejectsZero(t *testing.T) {
	tests := []struct {
		name    string
		setting string
		wantErr string
	}{
		{name: "zero shutdown timeout", setting: `shutdown_timeout: "0s"`, wantErr: "server.shutdown_timeout"},
		{name: "negative read timeout", setting: `read_timeout: "-5s"`, wantErr: "server.read_timeout"},
		{name: "unparseable write timeout", setting: `write_timeout: "soon"`, wantErr: "server.write_timeout"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadYAML(t, "server:\n  address: \":8080\"\n  "+tt.setting+"\n")
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load() error = %v, want it to name %s", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPSettingsDefaults(t *testing.T) {
	cfg, err := loadYAML(t, serverOnly)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	tests := []struct {
		name      string
		got, want httpx.Settings
	}{
		{
			name: "telegram",
			got:  cfg.Telegram.HTTP.Settings(),
			want: httpx.Settings{
				Timeout: 40 * time.Second, DialTimeout: 10 * time.Second, TLSHandshakeTimeout: 10 * time.Second,
				KeepAlive: 30 * time.Second, MaxIdleConns: 100, IdleConnTimeout: 90 * time.Second,
				ResponseHeaderTimeout: 35 * time.Second, ExpectContinueTimeout: time.Second,
			},
		},
		{
			name: "github",
			got:  cfg.GitHub.HTTP.Settings(),
			want: httpx.Settings{
				Timeout: 30 * time.Second, DialTimeout: 30 * time.Second, TLSHandshakeTimeout: 10 * time.Second,
				KeepAlive: 30 * time.Second, MaxIdleConns: 100, IdleConnTimeout: 90 * time.Second,
				EnvProxy: true, ForceHTTP2: true,
			},
		},
		{
			name: "kubernetes",
			got:  cfg.Kubernetes.HTTP.Settings(),
			want: httpx.Settings{
				Timeout: 15 * time.Second, DialTimeout: 10 * time.Second, TLSHandshakeTimeout: 10 * time.Second,
				KeepAlive: 30 * time.Second, MaxIdleConns: 100, IdleConnTimeout: 90 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s settings = %+v, want %+v", tt.name, tt.got, tt.want)
		}
	}
}

func TestKubernetesTimeoutOverride(t *testing.T) {
	cfg, err := loadYAML(t, serverOnly+"kubernetes:\n  http:\n    timeout: \"20s\"\n")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.Kubernetes.HTTP.Settings().Timeout; got != 20*time.Second {
		t.Errorf("kubernetes timeout = %s, want 20s", got)
	}
}

// TestProxyAndHTTP2OnlyForGitHub checks that only the GitHub settings keep the proxy and HTTP/2 of the default transport.
func TestProxyAndHTTP2OnlyForGitHub(t *testing.T) {
	cfg, err := loadYAML(t, serverOnly)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if s := cfg.GitHub.HTTP.Settings(); !s.EnvProxy || !s.ForceHTTP2 {
		t.Errorf("github EnvProxy, ForceHTTP2 = %v, %v; want true, true", s.EnvProxy, s.ForceHTTP2)
	}
	tests := []struct {
		name string
		s    httpx.Settings
	}{
		{name: "telegram", s: cfg.Telegram.HTTP.Settings()},
		{name: "kubernetes", s: cfg.Kubernetes.HTTP.Settings()},
	}
	for _, tt := range tests {
		if tt.s.EnvProxy || tt.s.ForceHTTP2 {
			t.Errorf("%s EnvProxy, ForceHTTP2 = %v, %v; want false, false", tt.name, tt.s.EnvProxy, tt.s.ForceHTTP2)
		}
	}
}

func TestHTTPSettingsRejectNonPositive(t *testing.T) {
	tests := []struct {
		name    string
		section string
		wantErr string
	}{
		{name: "github negative idle connections", section: "github:\n  http:\n    max_idle_conns: -1\n", wantErr: "github.http.max_idle_conns"},
		{name: "kubernetes zero timeout", section: "kubernetes:\n  http:\n    timeout: \"0s\"\n", wantErr: "kubernetes.http.timeout"},
		{name: "telegram zero dial timeout", section: "telegram:\n  http:\n    dial_timeout: \"0s\"\n", wantErr: "telegram.http.dial_timeout"},
		{name: "github negative idle timeout", section: "github:\n  http:\n    idle_conn_timeout: \"-1s\"\n", wantErr: "github.http.idle_conn_timeout"},
		{name: "kubernetes zero keep alive", section: "kubernetes:\n  http:\n    keep_alive: \"0s\"\n", wantErr: "kubernetes.http.keep_alive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadYAML(t, serverOnly+tt.section)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load() error = %v, want it to name %s", err, tt.wantErr)
			}
		})
	}
}

func TestAPIURLValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantURL string
		wantErr string
	}{
		{name: "default", content: serverOnly, wantURL: "https://api.github.com"},
		{name: "GHES with trailing slash", content: serverOnly + "github:\n  api_url: \"https://ghe.example.com/api/v3/\"\n", wantURL: "https://ghe.example.com/api/v3"},
		{name: "plain http rejected", content: serverOnly + "github:\n  api_url: \"http://ghe.example.com/api/v3\"\n", wantErr: "github.api_url"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loadYAML(t, tt.content)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Load() error = %v, want it to name %s", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.GitHub.APIURL != tt.wantURL {
				t.Errorf("APIURL = %q, want %q", cfg.GitHub.APIURL, tt.wantURL)
			}
		})
	}
}

func TestTelegramPollBoundsDerived(t *testing.T) {
	cfg, err := loadYAML(t, serverOnly+"telegram:\n  http:\n    poll_timeout: \"50s\"\n")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	h := cfg.Telegram.HTTP
	if h.PollTimeout != 50*time.Second || h.Timeout != 60*time.Second || h.ResponseHeaderTimeout != 55*time.Second {
		t.Errorf("poll, timeout, header = %s, %s, %s; want 50s, 60s, 55s", h.PollTimeout, h.Timeout, h.ResponseHeaderTimeout)
	}
}

func TestTelegramTimeoutNotAbovePoll(t *testing.T) {
	tests := []struct {
		name    string
		section string
		wantErr string
	}{
		{name: "timeout equal to poll", section: "poll_timeout: \"30s\"\n    timeout: \"30s\"", wantErr: "telegram.http.timeout"},
		{name: "header timeout equal to poll", section: "poll_timeout: \"30s\"\n    response_header_timeout: \"30s\"", wantErr: "telegram.http.response_header_timeout"},
		{name: "zero poll timeout", section: "poll_timeout: \"0s\"", wantErr: "telegram.http.poll_timeout"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadYAML(t, serverOnly+"telegram:\n  http:\n    "+tt.section+"\n")
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load() error = %v, want it to name %s", err, tt.wantErr)
			}
		})
	}
}

func TestRetryConfigDefaultsAndMinimum(t *testing.T) {
	cfg, err := loadYAML(t, serverOnly)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if h := cfg.Telegram.HTTP; h.MaxAttempts != 3 || h.RetryBackoff != time.Second || h.MaxRetryAfter != 30*time.Second {
		t.Errorf("retry = %d, %s, %s; want 3, 1s, 30s", h.MaxAttempts, h.RetryBackoff, h.MaxRetryAfter)
	}

	tests := []struct {
		name    string
		setting string
		wantErr string
	}{
		{name: "zero attempts", setting: "max_attempts: 0", wantErr: "telegram.http.max_attempts"},
		{name: "one attempt allowed", setting: "max_attempts: 1"},
		{name: "zero backoff", setting: `retry_backoff: "0s"`, wantErr: "telegram.http.retry_backoff"},
		{name: "negative max retry after", setting: `max_retry_after: "-1s"`, wantErr: "telegram.http.max_retry_after"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadYAML(t, serverOnly+"telegram:\n  http:\n    "+tt.setting+"\n")
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Load() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load() error = %v, want it to name %s", err, tt.wantErr)
			}
		})
	}
}

func TestMaxResponseBytesValidation(t *testing.T) {
	cfg, err := loadYAML(t, serverOnly)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GitHub.HTTP.MaxResponseBytes != httpx.DefaultMaxResponseBytes || cfg.Kubernetes.HTTP.MaxResponseBytes != httpx.DefaultMaxResponseBytes {
		t.Errorf("max_response_bytes = %d, %d; want %d", cfg.GitHub.HTTP.MaxResponseBytes, cfg.Kubernetes.HTTP.MaxResponseBytes, httpx.DefaultMaxResponseBytes)
	}

	tests := []struct {
		name    string
		section string
		wantErr string
	}{
		{name: "github zero", section: "github:\n  http:\n    max_response_bytes: 0\n", wantErr: "github.http.max_response_bytes"},
		{name: "kubernetes negative", section: "kubernetes:\n  http:\n    max_response_bytes: -1\n", wantErr: "kubernetes.http.max_response_bytes"},
		{name: "github custom value", section: "github:\n  http:\n    max_response_bytes: 1024\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadYAML(t, serverOnly+tt.section)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Load() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load() error = %v, want it to name %s", err, tt.wantErr)
			}
		})
	}
}

// httpKeys are the key paths of the HTTP settings in the spec, each of which the example must document.
var httpKeys = []string{
	"server.read_timeout", "server.write_timeout", "server.shutdown_timeout",
	"github.api_url", "github.http.timeout", "github.http.dial_timeout", "github.http.tls_handshake_timeout",
	"github.http.keep_alive", "github.http.max_idle_conns", "github.http.idle_conn_timeout", "github.http.max_response_bytes",
	"kubernetes.http.timeout", "kubernetes.http.dial_timeout", "kubernetes.http.tls_handshake_timeout",
	"kubernetes.http.keep_alive", "kubernetes.http.max_idle_conns", "kubernetes.http.idle_conn_timeout", "kubernetes.http.max_response_bytes",
	"telegram.http.timeout", "telegram.http.dial_timeout", "telegram.http.tls_handshake_timeout",
	"telegram.http.keep_alive", "telegram.http.max_idle_conns", "telegram.http.idle_conn_timeout",
	"telegram.http.poll_timeout", "telegram.http.response_header_timeout", "telegram.http.expect_continue_timeout",
	"telegram.http.max_attempts", "telegram.http.retry_backoff", "telegram.http.max_retry_after",
}

// commentBlock returns the comment lines that start at the line start, with their "# " prefix removed.
// The block ends at the first line that is not a comment.
func commentBlock(text, start string) string {
	var lines []string
	found := false
	for _, line := range strings.Split(text, "\n") {
		if !found {
			if line != start {
				continue
			}
			found = true
		}
		if !strings.HasPrefix(line, "# ") {
			break
		}
		lines = append(lines, strings.TrimPrefix(line, "# "))
	}
	return strings.Join(lines, "\n")
}

func TestConfigExampleDocumentsNewKeys(t *testing.T) {
	example, err := os.ReadFile(filepath.Join("..", "..", "configs", "config.yaml.example"))
	if err != nil {
		t.Fatalf("read example config: %v", err)
	}
	documented := viper.New()
	documented.SetConfigType("yaml")
	if err := documented.ReadConfig(strings.NewReader(commentBlock(string(example), "# server:"))); err != nil {
		t.Fatalf("parse the documented HTTP block: %v", err)
	}
	for _, key := range httpKeys {
		if !documented.IsSet(key) {
			t.Errorf("configs/config.yaml.example does not document %s", key)
		}
	}
}

// nestedYAML returns a YAML document that sets the dotted key path to value.
func nestedYAML(key, value string) string {
	var b strings.Builder
	parts := strings.Split(key, ".")
	for i, part := range parts {
		b.WriteString(strings.Repeat("  ", i) + part + ":")
		if i == len(parts)-1 {
			b.WriteString(" " + value)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// TestDurationKeysRequireUnit checks every duration key: a bare integer fails and names the key,
// and a quoted value with a unit passes the unit check.
func TestDurationKeysRequireUnit(t *testing.T) {
	for _, key := range durationKeys {
		t.Run(key, func(t *testing.T) {
			_, err := loadYAML(t, nestedYAML(key, "30"))
			if err == nil || !strings.Contains(err.Error(), key) || !strings.Contains(err.Error(), "a unit is required") {
				t.Errorf("bare 30: Load() error = %v, want it to name %s and say a unit is required", err, key)
			}
			_, err = loadYAML(t, nestedYAML(key, `"30s"`))
			if err != nil && strings.Contains(err.Error(), "a unit is required") {
				t.Errorf(`quoted "30s": Load() error = %v, want no unit error`, err)
			}
		})
	}
}
