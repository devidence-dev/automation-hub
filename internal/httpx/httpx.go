// Package httpx builds the outbound HTTP clients and sends the requests whose
// response bodies are capped. It uses only the standard library.
package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// DefaultMaxResponseBytes is the body cap of a response that the config leaves uncapped: 1 MiB.
const DefaultMaxResponseBytes int64 = 1 << 20

// Settings are the transport values of one client. A zero ResponseHeaderTimeout or
// ExpectContinueTimeout means no limit, and a nil TLSConfig means the system roots.
// EnvProxy routes requests through the proxy variables of the environment. ForceHTTP2 keeps
// HTTP/2 on although the transport sets a dialer or TLS config, which otherwise turns it off.
// Both are off unless a client needs them.
type Settings struct {
	Timeout               time.Duration // whole request, including the body read
	DialTimeout           time.Duration
	TLSHandshakeTimeout   time.Duration
	KeepAlive             time.Duration // TCP keep-alive period of the dialer
	MaxIdleConns          int
	IdleConnTimeout       time.Duration
	ResponseHeaderTimeout time.Duration
	ExpectContinueTimeout time.Duration
	TLSConfig             *tls.Config
	EnvProxy              bool
	ForceHTTP2            bool
}

// Request is one HTTP request. A nil Body sends no body.
type Request struct {
	Method string
	URL    string
	Header http.Header
	Body   []byte
}

// Response is the answer to a Request. Body holds at most the maxBytes given to Send.
type Response struct {
	StatusCode int
	Status     string
	Body       []byte
}

// NewClient returns a client with its own transport, built from s.
func NewClient(s Settings) *http.Client {
	tr := &http.Transport{
		DialContext:           newDialer(s).DialContext,
		TLSClientConfig:       s.TLSConfig,
		TLSHandshakeTimeout:   s.TLSHandshakeTimeout,
		ResponseHeaderTimeout: s.ResponseHeaderTimeout,
		ExpectContinueTimeout: s.ExpectContinueTimeout,
		MaxIdleConns:          s.MaxIdleConns,
		IdleConnTimeout:       s.IdleConnTimeout,
		ForceAttemptHTTP2:     s.ForceHTTP2,
	}
	if s.EnvProxy {
		tr.Proxy = http.ProxyFromEnvironment
	}
	return &http.Client{Timeout: s.Timeout, Transport: tr}
}

// newDialer builds the dialer of NewClient from the dial timeout and keep-alive period of s.
func newDialer(s Settings) *net.Dialer {
	return &net.Dialer{Timeout: s.DialTimeout, KeepAlive: s.KeepAlive}
}

// Send sends req with client and reads the response body up to maxBytes. The status is
// returned as it is: the caller decides what a status means.
func Send(ctx context.Context, client *http.Client, req Request, maxBytes int64) (Response, error) {
	var body io.Reader
	if req.Body != nil {
		body = bytes.NewReader(req.Body)
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		return Response{}, fmt.Errorf("create request: %w", err)
	}
	for key, values := range req.Header {
		httpReq.Header[key] = values
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return Response{}, fmt.Errorf("read response: %w", err)
	}
	return Response{StatusCode: resp.StatusCode, Status: resp.Status, Body: data}, nil
}
