package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"automation-hub/internal/config"
)

const (
	testForkRepo = "devidence-dev/discord-tts-bot"
	// testForkTag is kept literal here so the test checks the spec text, not the constant.
	testForkTag = " (fork, solo informativo)" //nolint:misspell // Spanish text fixed by the spec
)

func loadFixture(t *testing.T) renovateReport {
	t.Helper()
	data, err := os.ReadFile("testdata/renovate-report.json")
	if err != nil {
		t.Fatal(err)
	}
	rep, err := parseReport(data)
	if err != nil {
		t.Fatalf("parseReport() error = %v", err)
	}
	return rep
}

// messageFor returns the message whose header names repo.
func messageFor(t *testing.T, messages []string, repo string) string {
	t.Helper()
	for _, msg := range messages {
		if strings.HasPrefix(msg, "<b>📦 "+repo+"</b>") {
			return msg
		}
	}
	t.Fatalf("no message for %s in %q", repo, messages)
	return ""
}

func TestBuildMessagesFixture(t *testing.T) {
	messages := buildMessages(loadFixture(t), []string{testForkRepo})

	// digest-only and warn-only repos produce nothing: 3 of the 5 repos remain.
	if len(messages) != 3 {
		t.Fatalf("messages = %d, want 3: %q", len(messages), messages)
	}
	for _, msg := range messages {
		if len(msg) > telegramLimit {
			t.Errorf("message has %d bytes, over the limit", len(msg))
		}
		if strings.Contains(msg, "digest-only") || strings.Contains(msg, "warn-only") {
			t.Errorf("repo without listed deps or errors produced a message: %q", msg)
		}
	}

	t.Run("major update", func(t *testing.T) {
		want := strings.Join([]string{
			"<b>📦 devidence-dev/QuantWarden</b>",
			"3/20 desactualizadas · 2.3 libyears",
			"⚠️ 2 errores en el scan",
			"",
			"🔴 <code>github.com/gorilla/mux</code> 1.8.0 → 2.0.0 (major), 1.9.0 (minor)",
			"🔴 <code>golang.org/x/net</code> 0.30.0 → 1.0.0 (major)",
			"🟡 <code>go.uber.org/zap</code> 1.27.0 → 1.27.1 (patch)",
		}, "\n")
		if got := messageFor(t, messages, "devidence-dev/QuantWarden"); got != want {
			t.Errorf("message =\n%s\nwant\n%s", got, want)
		}
	})

	t.Run("fork", func(t *testing.T) {
		want := "<b>📦 " + testForkRepo + "</b>" + testForkTag + "\n0/1 desactualizadas"
		got := messageFor(t, messages, testForkRepo)
		if !strings.HasPrefix(got, "<b>📦 "+testForkRepo+"</b>"+testForkTag+"\n") || !strings.Contains(got, "\n\n🟡 <code>discord.js</code> 14.1.0 → 14.2.0 (minor)") {
			t.Errorf("message = %q, want fork header and a yellow line (%s...)", got, want)
		}
	})

	t.Run("not a fork when not configured", func(t *testing.T) {
		got := messageFor(t, buildMessages(loadFixture(t), nil), testForkRepo)
		if strings.Contains(got, "fork") {
			t.Errorf("message = %q, want no fork tag", got)
		}
	})

	t.Run("more than 40 lines", func(t *testing.T) {
		got := messageFor(t, messages, "devidence-dev/big-monorepo")
		lines := strings.Split(got, "\n")
		// header, libyears, blank, 40 deps, "… y 5 más"
		if len(lines) != 3+40+1 || lines[len(lines)-1] != "… y 5 más" {
			t.Errorf("got %d lines, last %q", len(lines), lines[len(lines)-1])
		}
		if lines[1] != "45/60 desactualizadas · 10 libyears" {
			t.Errorf("libyears line = %q", lines[1])
		}
	})
}

func TestBuildMessagesUpToDate(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{name: "no repositories", json: `{"problems":[],"repositories":{}}`},
		{name: "digest only", json: `{"repositories":{"a/b":{"packageFiles":{"x":[{"deps":[{"depName":"d","updates":[{"updateType":"digest","newValue":"1"}]}]}]}}}}`},
		{name: "warnings only", json: `{"repositories":{"a/b":{"problems":[{"level":40}]}}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := parseReport([]byte(tt.json))
			if err != nil {
				t.Fatal(err)
			}
			got := buildMessages(rep, nil)
			if len(got) != 1 || got[0] != "✅ Dependency report: todo al día." {
				t.Errorf("messages = %q", got)
			}
		})
	}
}

func TestBuildMessagesErrorsWithoutDeps(t *testing.T) {
	rep, _ := parseReport([]byte(`{"repositories":{"a/b":{"problems":[{"level":50}]}}}`))
	got := buildMessages(rep, nil)
	if len(got) != 1 || got[0] != "<b>📦 a/b</b>\n⚠️ 1 errores en el scan" {
		t.Errorf("messages = %q", got)
	}
}

func TestBuildMessagesFallbacksAndEscaping(t *testing.T) {
	rep, err := parseReport([]byte(`{"repositories":{"o/r<x>":{"packageFiles":{"m":[{"deps":[
		{"updates":[{"updateType":"minor","newValue":"1.0 & 2"}]},
		{"depName":"a<b>","currentValue":"^1","updates":[{"updateType":"patch","newVersion":"1.1"}]}
	]}]}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	got := buildMessages(rep, nil)
	want := "<b>📦 o/r&lt;x&gt;</b>\n\n" +
		"🟡 <code>?</code> ? → 1.0 &amp; 2 (minor)\n" +
		"🟡 <code>a&lt;b&gt;</code> ^1 → 1.1 (patch)"
	if len(got) != 1 || got[0] != want {
		t.Errorf("messages = %q, want %q", got, want)
	}
}

func TestBuildMessagesLongEscapedContentIsCutByWholeLines(t *testing.T) {
	var deps []string
	for i := 0; i < 40; i++ {
		deps = append(deps, fmt.Sprintf(`{"depName":"<<<&&&%02d%s>>>","currentValue":"1","updates":[{"updateType":"minor","newValue":"2"}]}`, i, strings.Repeat("&<", 30)))
	}
	rep, err := parseReport([]byte(`{"repositories":{"a/b":{"packageFiles":{"m":[{"deps":[` + strings.Join(deps, ",") + `]}]}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	messages := buildMessages(rep, nil)
	if len(messages) != 1 {
		t.Fatalf("messages = %d", len(messages))
	}
	msg := messages[0]
	if len(msg) > telegramLimit {
		t.Fatalf("message has %d bytes", len(msg))
	}
	lines := strings.Split(msg, "\n")
	last := lines[len(lines)-1]
	if !strings.HasPrefix(last, "… y ") || !strings.HasSuffix(last, " más") {
		t.Errorf("last line = %q, want the remainder line", last)
	}
	for _, line := range lines[4 : len(lines)-1] {
		if !strings.HasPrefix(line, "🟡 <code>") || !strings.Contains(line, "</code> 1 → 2 (minor)") {
			t.Errorf("line was cut inside a tag or entity: %q", line)
		}
	}
	// No raw "<" or "&" may appear outside the known tags and entities.
	plain := strings.NewReplacer("<code>", "", "</code>", "", "<b>", "", "</b>", "", "&lt;", "", "&gt;", "", "&amp;", "").Replace(msg)
	if strings.ContainsAny(plain, "<>&") {
		t.Errorf("unescaped markup in message")
	}
}

func TestBuildFailureMessage(t *testing.T) {
	const pod = "dependency-report-manual-1791763200-x7k2p"
	t.Run("escapes the tail", func(t *testing.T) {
		got := buildFailureMessage(1, pod, "<html>\nfoo & bar\n")
		want := "❌ Renovate falló (exit 1) en el pod " + pod + "\n<pre>&lt;html&gt;\nfoo &amp; bar</pre>"
		if got != want {
			t.Errorf("message = %q, want %q", got, want)
		}
	})
	t.Run("keeps the last lines that fit", func(t *testing.T) {
		var lines []string
		for i := 0; i < 400; i++ {
			lines = append(lines, fmt.Sprintf("line %03d <err> & more text here", i))
		}
		got := buildFailureMessage(2, pod, strings.Join(lines, "\n")) // about 12 KB
		if len(got) > telegramLimit {
			t.Fatalf("message has %d bytes", len(got))
		}
		if !strings.Contains(got, "line 399 &lt;err&gt; &amp; more text here</pre>") || strings.Contains(got, "line 000") {
			t.Errorf("tail not cut from the front: %q", got[:200])
		}
		body := got[strings.Index(got, "<pre>")+5 : len(got)-len("</pre>")]
		if first := strings.SplitN(body, "\n", 2)[0]; !strings.HasPrefix(first, "line ") {
			t.Errorf("first kept line is partial: %q", first)
		}
	})
	t.Run("one huge line is cut without splitting an entity", func(t *testing.T) {
		got := buildFailureMessage(1, pod, strings.Repeat("<", 10000))
		if len(got) > telegramLimit {
			t.Fatalf("message has %d bytes", len(got))
		}
		body := got[strings.Index(got, "<pre>")+5 : len(got)-len("</pre>")]
		if strings.ReplaceAll(body, "&lt;", "") != "" {
			t.Errorf("partial entity in %q", body[:20])
		}
	})
	t.Run("empty tail", func(t *testing.T) {
		if got := buildFailureMessage(1, pod, ""); got != "❌ Renovate falló (exit 1) en el pod "+pod {
			t.Errorf("message = %q", got)
		}
	})
}

// fakeHTMLSender records messages and can fail every send.
type fakeHTMLSender struct {
	mu       sync.Mutex
	chatIDs  []string
	messages []string
	err      error
}

func (f *fakeHTMLSender) SendHTML(chatID, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.chatIDs = append(f.chatIDs, chatID)
	f.messages = append(f.messages, text)
	return f.err
}

func (f *fakeHTMLSender) waitFor(t *testing.T, n int) []string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		f.mu.Lock()
		got := append([]string(nil), f.messages...)
		f.mu.Unlock()
		if len(got) >= n {
			return got
		}
		if time.Now().After(deadline) {
			t.Fatalf("messages = %q after 2s, want at least %d", got, n)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

type reportRequest struct {
	token  string
	fields map[string]string
	files  map[string]string
}

func (rr reportRequest) build(t *testing.T) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range rr.fields {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for k, v := range rr.files {
		fw, err := mw.CreateFormFile(k, k)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fw.Write([]byte(v))
	}
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/webhook/dependency-report", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if rr.token != "" {
		req.Header.Set("X-Report-Token", rr.token)
	}
	return req
}

func newReportHandler(sender htmlSender) *DependencyReportHandler {
	return NewDependencyReportHandler(config.DependencyReportConfig{
		WebhookToken: "secret", TelegramChatID: "555", Forks: []string{testForkRepo},
	}, sender, zap.NewNop())
}

func TestDependencyReportHandler(t *testing.T) {
	fixture, err := os.ReadFile("testdata/renovate-report.json")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name         string
		req          reportRequest
		wantStatus   int
		wantMessages int
		check        func(t *testing.T, messages []string)
	}{
		{name: "missing token", req: reportRequest{fields: map[string]string{"exit_code": "0"}, files: map[string]string{"report": string(fixture)}}, wantStatus: http.StatusUnauthorized},
		{name: "wrong token", req: reportRequest{token: "nope", fields: map[string]string{"exit_code": "0"}, files: map[string]string{"report": string(fixture)}}, wantStatus: http.StatusUnauthorized},
		{name: "non-integer exit code", req: reportRequest{token: "secret", fields: map[string]string{"exit_code": "abc"}}, wantStatus: http.StatusBadRequest},
		{name: "missing exit code", req: reportRequest{token: "secret", fields: map[string]string{"pod": "p"}}, wantStatus: http.StatusBadRequest},
		{name: "invalid report json", req: reportRequest{token: "secret", fields: map[string]string{"exit_code": "0"}, files: map[string]string{"report": "{not json"}}, wantStatus: http.StatusBadRequest},
		{
			name: "valid report; forks come from config, not the request",
			req: reportRequest{
				token:  "secret",
				fields: map[string]string{"exit_code": "0", "pod": "p", "forks": `["devidence-dev/QuantWarden"]`},
				files:  map[string]string{"report": string(fixture), "log_tail": "ignored"},
			},
			wantStatus: http.StatusAccepted, wantMessages: 3,
			check: func(t *testing.T, messages []string) {
				if !strings.Contains(messageFor(t, messages, testForkRepo), testForkTag) {
					t.Error("configured fork not tagged")
				}
				if strings.Contains(messageFor(t, messages, "devidence-dev/QuantWarden"), "fork") {
					t.Error("a forks field in the request must be ignored")
				}
			},
		},
		{
			name:       "renovate failure",
			req:        reportRequest{token: "secret", fields: map[string]string{"exit_code": "1", "pod": "pod-x"}, files: map[string]string{"log_tail": "<html>"}},
			wantStatus: http.StatusAccepted, wantMessages: 1,
			check: func(t *testing.T, messages []string) {
				if !strings.HasPrefix(messages[0], "❌ Renovate falló (exit 1) en el pod pod-x") || !strings.Contains(messages[0], "&lt;html&gt;") {
					t.Errorf("message = %q", messages[0])
				}
			},
		},
		{
			name:       "exit 0 without report is a failure",
			req:        reportRequest{token: "secret", fields: map[string]string{"exit_code": "0", "pod": "pod-x"}},
			wantStatus: http.StatusAccepted, wantMessages: 1,
			check: func(t *testing.T, messages []string) {
				if !strings.HasPrefix(messages[0], "❌ Renovate falló (exit 0)") {
					t.Errorf("message = %q", messages[0])
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeHTMLSender{}
			rec := httptest.NewRecorder()
			newReportHandler(sender).ServeHTTP(rec, tt.req.build(t))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantMessages == 0 {
				time.Sleep(20 * time.Millisecond) // give a wrong background send a chance to show up
				if got := sender.waitForNone(); len(got) != 0 {
					t.Fatalf("messages sent = %q, want none", got)
				}
				return
			}
			messages := sender.waitFor(t, tt.wantMessages)
			if len(messages) != tt.wantMessages {
				t.Fatalf("messages = %d, want %d", len(messages), tt.wantMessages)
			}
			for _, chatID := range sender.chatIDs {
				if chatID != "555" {
					t.Errorf("chat = %q, want the configured one", chatID)
				}
			}
			if tt.check != nil {
				tt.check(t, messages)
			}
		})
	}
}

func (f *fakeHTMLSender) waitForNone() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.messages...)
}

func TestDependencyReportHandlerRejectsOversizedBody(t *testing.T) {
	sender := &fakeHTMLSender{}
	rec := httptest.NewRecorder()
	req := reportRequest{token: "secret", fields: map[string]string{"exit_code": "1"}, files: map[string]string{"log_tail": strings.Repeat("x", maxReportBody+1)}}.build(t)
	newReportHandler(sender).ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
	if got := sender.waitForNone(); len(got) != 0 {
		t.Errorf("messages sent = %q", got)
	}
}

func TestDependencyReportHandlerNotMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook/dependency-report", strings.NewReader("exit_code=0"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Report-Token", "secret")
	rec := httptest.NewRecorder()
	newReportHandler(&fakeHTMLSender{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestDependencyReportHandlerTelegramFailureDoesNotPanicOrBlock(t *testing.T) {
	sender := &fakeHTMLSender{err: errors.New("telegram down")}
	rec := httptest.NewRecorder()
	req := reportRequest{token: "secret", fields: map[string]string{"exit_code": "1"}, files: map[string]string{"log_tail": "x"}}.build(t)
	newReportHandler(sender).ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 even when Telegram fails", rec.Code)
	}
	sender.waitFor(t, 1) // the failed send was attempted and only logged
}
