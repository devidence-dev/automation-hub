package httpx

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClientAppliesSettings(t *testing.T) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	s := Settings{
		Timeout:               20 * time.Second,
		DialTimeout:           3 * time.Second,
		TLSHandshakeTimeout:   4 * time.Second,
		KeepAlive:             5 * time.Second,
		MaxIdleConns:          7,
		IdleConnTimeout:       8 * time.Second,
		ResponseHeaderTimeout: 9 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSConfig:             tlsConfig,
		EnvProxy:              true,
		ForceHTTP2:            true,
	}
	client := NewClient(s)
	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport = %T, want *http.Transport", client.Transport)
	}
	dialer := newDialer(s)

	tests := []struct {
		name      string
		got, want any
	}{
		{"client timeout", client.Timeout, s.Timeout},
		{"dial timeout", dialer.Timeout, s.DialTimeout},
		{"keep alive", dialer.KeepAlive, s.KeepAlive},
		{"TLS handshake timeout", tr.TLSHandshakeTimeout, s.TLSHandshakeTimeout},
		{"max idle conns", tr.MaxIdleConns, s.MaxIdleConns},
		{"idle conn timeout", tr.IdleConnTimeout, s.IdleConnTimeout},
		{"response header timeout", tr.ResponseHeaderTimeout, s.ResponseHeaderTimeout},
		{"expect continue timeout", tr.ExpectContinueTimeout, s.ExpectContinueTimeout},
		{"TLS config", tr.TLSClientConfig, tlsConfig},
		{"proxy from environment", tr.Proxy != nil, true},
		{"dial context", tr.DialContext != nil, true},
		{"HTTP/2 forced", tr.ForceAttemptHTTP2, true},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

func TestNewClientWithoutEnvProxyOrHTTP2(t *testing.T) {
	client := NewClient(Settings{Timeout: time.Second})
	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport = %T, want *http.Transport", client.Transport)
	}
	if tr.Proxy != nil {
		t.Error("Proxy is set, want no proxy when EnvProxy is false")
	}
	if tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = true, want false when ForceHTTP2 is false")
	}
}

func TestSendCapsBodyAtLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, strings.Repeat("x", 2000))
	}))
	defer srv.Close()

	resp, err := Send(context.Background(), srv.Client(), Request{Method: http.MethodGet, URL: srv.URL}, 1024)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if len(resp.Body) != 1024 {
		t.Errorf("body length = %d, want 1024", len(resp.Body))
	}
}

func TestSendReturnsStatusAndBody(t *testing.T) {
	var gotMethod, gotHeader, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotHeader = r.Header.Get("X-Test")
		payload, _ := io.ReadAll(r.Body)
		gotBody = string(payload)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "bad gateway")
	}))
	defer srv.Close()

	req := Request{
		Method: http.MethodPost,
		URL:    srv.URL,
		Header: http.Header{"X-Test": {"1"}},
		Body:   []byte(`{"ref":"main"}`),
	}
	resp, err := Send(context.Background(), srv.Client(), req, DefaultMaxResponseBytes)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if resp.StatusCode != http.StatusBadGateway || resp.Status != "502 Bad Gateway" {
		t.Errorf("status = %d %q, want 502 %q", resp.StatusCode, resp.Status, "502 Bad Gateway")
	}
	if string(resp.Body) != "bad gateway" {
		t.Errorf("body = %q, want %q", resp.Body, "bad gateway")
	}
	if gotMethod != http.MethodPost || gotHeader != "1" || gotBody != `{"ref":"main"}` {
		t.Errorf("server saw method %q header %q body %q", gotMethod, gotHeader, gotBody)
	}
}
