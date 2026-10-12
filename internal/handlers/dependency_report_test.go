package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

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

// assertFixtureMessages checks the message count, the size limit and that the
// digest-only and warn-only repos produced nothing.
func assertFixtureMessages(t *testing.T, messages []string) {
	t.Helper()
	// digest-only and warn-only repos produce nothing: 4 of the 6 repos remain.
	if len(messages) != 4 {
		t.Fatalf("messages = %d, want 4: %q", len(messages), messages)
	}
	for _, msg := range messages {
		if len(msg) > telegramLimit {
			t.Errorf("message has %d bytes, over the limit", len(msg))
		}
		if strings.Contains(msg, "digest-only") || strings.Contains(msg, "warn-only") {
			t.Errorf("repo without listed deps or errors produced a message: %q", msg)
		}
	}
}

// assertTruncatedMessage checks a message cut at the 40-line cap.
func assertTruncatedMessage(t *testing.T, got string) {
	t.Helper()
	lines := strings.Split(got, "\n")
	// header, libyears, blank, 40 deps, "… y 5 más"
	if len(lines) != 3+40+1 || lines[len(lines)-1] != "… y 5 más" {
		t.Errorf("got %d lines, last %q", len(lines), lines[len(lines)-1])
	}
	if lines[1] != "45/60 desactualizadas · 10 libyears" {
		t.Errorf("libyears line = %q", lines[1])
	}
}

func TestBuildMessagesFixture(t *testing.T) {
	messages := buildMessages(loadFixture(t), []string{testForkRepo})
	assertFixtureMessages(t, messages)

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

	t.Run("homelab escapes names and keeps datasource out of the text", func(t *testing.T) {
		want := strings.Join([]string{
			"<b>📦 devidence-dev/homelab</b>",
			"",
			"🟡 <code>&lt;script&gt;alert(1)&lt;/script&gt;</code> 1.0.0 → 1.0.1 (patch)",
			"🟡 <code>actions/checkout</code> v4.2.0 → v4.2.2 (patch)",
			"🟡 <code>ghcr.io/immich-app/immich-server</code> v1.118.0 → v1.119.0 (minor)",
			"🟡 <code>legacy-tool</code> 1.0.0 → 1.0.1 (patch)",
		}, "\n")
		if got := messageFor(t, messages, "devidence-dev/homelab"); got != want {
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
		assertTruncatedMessage(t, messageFor(t, messages, "devidence-dev/big-monorepo"))
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

// assertKeepsLastLines checks that a long log keeps its newest lines whole.
func assertKeepsLastLines(t *testing.T, pod string) {
	t.Helper()
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
		assertKeepsLastLines(t, pod)
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

// sentDocument is one document send recorded by fakeHTMLSender.
type sentDocument struct {
	chatID, filename string
	data             []byte
}

// fakeHTMLSender records calls in order. err fails message sends (only the
// first failFirst of them when failFirst > 0); docErr fails document sends.
type fakeHTMLSender struct {
	mu        sync.Mutex
	order     []string // "message" or "document", in call order
	chatIDs   []string
	messages  []string
	docs      []sentDocument
	err       error
	failFirst int
	docErr    error
}

func (f *fakeHTMLSender) SendHTML(chatID, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.order = append(f.order, "message")
	f.chatIDs = append(f.chatIDs, chatID)
	f.messages = append(f.messages, text)
	if f.err != nil && (f.failFirst == 0 || len(f.messages) <= f.failFirst) {
		return f.err
	}
	return nil
}

func (f *fakeHTMLSender) SendDocument(chatID, filename string, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.order = append(f.order, "document")
	f.chatIDs = append(f.chatIDs, chatID)
	f.docs = append(f.docs, sentDocument{chatID: chatID, filename: filename, data: data})
	return f.docErr
}

// waitForCalls waits until n calls (messages and documents) were recorded and returns the order.
func (f *fakeHTMLSender) waitForCalls(t *testing.T, n int) []string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		f.mu.Lock()
		got := append([]string(nil), f.order...)
		f.mu.Unlock()
		if len(got) >= n {
			return got
		}
		if time.Now().After(deadline) {
			t.Fatalf("calls = %q after 2s, want at least %d", got, n)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func (f *fakeHTMLSender) waitFor(t *testing.T, n int) []string {
	t.Helper()
	f.waitForCalls(t, n)
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.messages...)
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

func newReportHandler(sender reportSender) *DependencyReportHandler {
	return NewDependencyReportHandler(config.DependencyReportConfig{
		WebhookToken: "secret", TelegramChatID: "555", Forks: []string{testForkRepo},
	}, sender, zap.NewNop())
}

// fixedClock is the time the hub "receives" the report in the handler tests: 2026-10-12 09:07 Guayaquil.
var fixedClock = time.Date(2026, 10, 12, 14, 7, 0, 0, time.UTC)

func newReportHandlerAt(sender reportSender, now time.Time, logger *zap.Logger) *DependencyReportHandler {
	h := NewDependencyReportHandler(config.DependencyReportConfig{
		WebhookToken: "secret", TelegramChatID: "555", Forks: []string{testForkRepo},
	}, sender, logger)
	h.now = func() time.Time { return now }
	return h
}

// reportHandlerCase is one request to the dependency report handler and what it must answer and send.
type reportHandlerCase struct {
	name         string
	req          reportRequest
	wantStatus   int
	wantMessages int
	check        func(t *testing.T, messages []string)
}

// runReportHandlerCase posts the request and checks the status and the messages sent.
func runReportHandlerCase(t *testing.T, tt reportHandlerCase) {
	t.Helper()
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
}

func TestDependencyReportHandler(t *testing.T) {
	fixture, err := os.ReadFile("testdata/renovate-report.json")
	if err != nil {
		t.Fatal(err)
	}
	tests := []reportHandlerCase{
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
			wantStatus: http.StatusAccepted, wantMessages: 4,
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
			runReportHandlerCase(t, tt)
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

// findDep returns the dep named depName under one package manager of a repository.
func findDep(t *testing.T, repo repoReport, manager, depName string) reportDep {
	t.Helper()
	for _, file := range repo.PackageFiles[manager] {
		for _, dep := range file.Deps {
			if dep.DepName != nil && *dep.DepName == depName {
				return dep
			}
		}
	}
	t.Fatalf("dep %s not found under %s", depName, manager)
	return reportDep{}
}

func TestParseReportDatasource(t *testing.T) {
	rep := loadFixture(t)
	homelab := rep.Repositories["devidence-dev/homelab"]

	t.Run("docker dep", func(t *testing.T) {
		dep := findDep(t, homelab, "helm-values", "ghcr.io/immich-app/immich-server")
		if dep.Datasource == nil || *dep.Datasource != "docker" {
			t.Errorf("datasource = %v, want docker", dep.Datasource)
		}
	})
	t.Run("missing datasource is nil", func(t *testing.T) {
		if dep := findDep(t, homelab, "ansible-galaxy", "legacy-tool"); dep.Datasource != nil {
			t.Errorf("datasource = %q, want nil", *dep.Datasource)
		}
	})
	t.Run("package file name", func(t *testing.T) {
		files := homelab.PackageFiles["helm-values"]
		if len(files) != 1 || files[0].PackageFile != "kubernetes/apps/immich/values.yaml" {
			t.Errorf("package files = %+v", files)
		}
	})
}

func postReport(t *testing.T, h *DependencyReportHandler, req reportRequest) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.build(t))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 (%s)", rec.Code, rec.Body.String())
	}
}

func validReport(t *testing.T) reportRequest {
	t.Helper()
	fixture, err := os.ReadFile("testdata/renovate-report.json")
	if err != nil {
		t.Fatal(err)
	}
	return reportRequest{token: "secret", fields: map[string]string{"exit_code": "0", "pod": "p"}, files: map[string]string{"report": string(fixture)}}
}

func TestDeliverMessagesThenDocument(t *testing.T) {
	sender := &fakeHTMLSender{}
	h := newReportHandlerAt(sender, fixedClock, zap.NewNop())
	postReport(t, h, validReport(t))

	order := sender.waitForCalls(t, 5)
	want := []string{"message", "message", "message", "message", "document"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("call order = %q, want %q", order, want)
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if len(sender.docs) != 1 {
		t.Fatalf("documents = %d, want 1", len(sender.docs))
	}
	doc := sender.docs[0]
	if doc.filename != "dependency-report-2026-10-12.html" || doc.chatID != "555" {
		t.Errorf("document = %q to chat %q", doc.filename, doc.chatID)
	}
	if !strings.Contains(string(doc.data), "2026-10-12 09:07 America/Guayaquil") {
		t.Error("document header lacks the Guayaquil scan time")
	}
}

func TestNoDocumentOnRenovateFailure(t *testing.T) {
	tests := []struct {
		name string
		req  reportRequest
	}{
		{name: "exit code 1 with a report", req: reportRequest{token: "secret", fields: map[string]string{"exit_code": "1", "pod": "p"}, files: map[string]string{"report": "{}", "log_tail": "x"}}},
		{name: "exit code 0 without a report", req: reportRequest{token: "secret", fields: map[string]string{"exit_code": "0", "pod": "p"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeHTMLSender{}
			postReport(t, newReportHandlerAt(sender, fixedClock, zap.NewNop()), tt.req)
			if got := sender.waitFor(t, 1); len(got) != 1 {
				t.Fatalf("messages = %q, want one failure message", got)
			}
			time.Sleep(20 * time.Millisecond) // give a wrong document send a chance to show up
			if docs := sender.waitForDocs(); len(docs) != 0 {
				t.Errorf("documents sent = %d, want none", len(docs))
			}
		})
	}
}

func TestDocumentWhenAllUpToDate(t *testing.T) {
	sender := &fakeHTMLSender{}
	req := reportRequest{token: "secret", fields: map[string]string{"exit_code": "0", "pod": "p"}, files: map[string]string{"report": `{"repositories":{}}`}}
	postReport(t, newReportHandlerAt(sender, fixedClock, zap.NewNop()), req)

	order := sender.waitForCalls(t, 2)
	if !reflect.DeepEqual(order, []string{"message", "document"}) {
		t.Fatalf("call order = %q", order)
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if sender.messages[0] != "✅ Dependency report: todo al día." {
		t.Errorf("message = %q", sender.messages[0])
	}
	if sender.docs[0].filename != "dependency-report-2026-10-12.html" {
		t.Errorf("document = %q", sender.docs[0].filename)
	}
}

func TestDocumentFailureIsLogged(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	sender := &fakeHTMLSender{docErr: errors.New("telegram rejected the document")}
	postReport(t, newReportHandlerAt(sender, fixedClock, zap.New(core)), validReport(t))

	sender.waitForCalls(t, 5)
	deadline := time.Now().Add(2 * time.Second)
	for logs.FilterMessage("Failed to send dependency report document").Len() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("document failure was not logged as an error")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if got := len(sender.waitFor(t, 4)); got != 4 {
		t.Errorf("messages recorded = %d, want 4", got)
	}
}

func TestMessageFailureStillSendsDocument(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	sender := &fakeHTMLSender{err: errors.New("telegram down"), failFirst: 1}
	postReport(t, newReportHandlerAt(sender, fixedClock, zap.New(core)), validReport(t))

	order := sender.waitForCalls(t, 5)
	if !reflect.DeepEqual(order, []string{"message", "message", "message", "message", "document"}) {
		t.Fatalf("call order = %q, want every message and the document", order)
	}
	if logs.FilterMessage("Failed to send dependency report message").Len() != 1 {
		t.Errorf("failed message not logged once")
	}
	if docs := sender.waitForDocs(); len(docs) != 1 {
		t.Errorf("documents = %d, want 1", len(docs))
	}
}

// waitForDocs returns the documents recorded so far.
func (f *fakeHTMLSender) waitForDocs() []sentDocument {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]sentDocument(nil), f.docs...)
}
