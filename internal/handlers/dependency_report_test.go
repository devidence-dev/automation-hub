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

// sentCall is one send recorded by fakeSender.
type sentCall struct {
	kind     string // "message" or "document"
	chatID   string
	text     string // the message text, or the caption of a document
	filename string
	data     []byte
}

// fakeSender records the sends in call order. err fails SendHTML and docErr fails SendDocument.
type fakeSender struct {
	mu     sync.Mutex
	calls  []sentCall
	err    error
	docErr error
}

func (f *fakeSender) SendHTML(chatID, text string) error {
	f.record(sentCall{kind: "message", chatID: chatID, text: text})
	return f.err
}

func (f *fakeSender) SendDocument(chatID, filename, caption string, data []byte) error {
	f.record(sentCall{kind: "document", chatID: chatID, text: caption, filename: filename, data: data})
	return f.docErr
}

func (f *fakeSender) record(c sentCall) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, c)
}

// snapshot returns the calls recorded so far.
func (f *fakeSender) snapshot() []sentCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]sentCall(nil), f.calls...)
}

// waitForCalls waits until at least n calls were recorded and returns them.
func (f *fakeSender) waitForCalls(t *testing.T, n int) []sentCall {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		calls := f.snapshot()
		if len(calls) >= n {
			return calls
		}
		if time.Now().After(deadline) {
			t.Fatalf("calls = %d after 2s, want at least %d", len(calls), n)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// kinds returns the kind of each call, in order.
func kinds(calls []sentCall) []string {
	out := make([]string, 0, len(calls))
	for _, c := range calls {
		out = append(out, c.kind)
	}
	return out
}

// onlyDocument returns the single call, which must be a document.
func onlyDocument(t *testing.T, calls []sentCall) sentCall {
	t.Helper()
	if len(calls) != 1 || calls[0].kind != "document" {
		t.Fatalf("calls = %q, want one document", kinds(calls))
	}
	return calls[0]
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
	name       string
	req        reportRequest
	wantStatus int
	wantCalls  int // sends expected after a 202
	check      func(t *testing.T, calls []sentCall)
}

// runReportHandlerCase posts the request and checks the status and the sends.
func runReportHandlerCase(t *testing.T, tt reportHandlerCase) {
	t.Helper()
	sender := &fakeSender{}
	rec := httptest.NewRecorder()
	newReportHandler(sender).ServeHTTP(rec, tt.req.build(t))
	if rec.Code != tt.wantStatus {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, tt.wantStatus, rec.Body.String())
	}
	if tt.wantCalls == 0 {
		time.Sleep(20 * time.Millisecond) // give a wrong background send a chance to show up
	}
	calls := sender.waitForCalls(t, tt.wantCalls)
	if len(calls) != tt.wantCalls {
		t.Fatalf("calls = %q, want %d", kinds(calls), tt.wantCalls)
	}
	for _, c := range calls {
		if c.chatID != "555" {
			t.Errorf("chat = %q, want the configured one", c.chatID)
		}
	}
	if tt.check != nil {
		tt.check(t, calls)
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
			wantStatus: http.StatusAccepted, wantCalls: 1,
			check: func(t *testing.T, calls []sentCall) {
				caption := onlyDocument(t, calls).text
				if !strings.Contains(caption, "discord-tts-bot (fork) — ") {
					t.Error("configured fork not tagged in the caption")
				}
				if strings.Contains(caption, "QuantWarden (fork)") {
					t.Error("a forks field in the request must be ignored")
				}
			},
		},
		{
			name:       "renovate failure",
			req:        reportRequest{token: "secret", fields: map[string]string{"exit_code": "1", "pod": "pod-x"}, files: map[string]string{"log_tail": "<html>"}},
			wantStatus: http.StatusAccepted, wantCalls: 1,
			check: func(t *testing.T, calls []sentCall) {
				if calls[0].kind != "message" || !strings.HasPrefix(calls[0].text, "❌ Renovate falló (exit 1) en el pod pod-x") || !strings.Contains(calls[0].text, "&lt;html&gt;") {
					t.Errorf("call = %+v", calls[0])
				}
			},
		},
		{
			name:       "exit 0 without report is a failure",
			req:        reportRequest{token: "secret", fields: map[string]string{"exit_code": "0", "pod": "pod-x"}},
			wantStatus: http.StatusAccepted, wantCalls: 1,
			check: func(t *testing.T, calls []sentCall) {
				if calls[0].kind != "message" || !strings.HasPrefix(calls[0].text, "❌ Renovate falló (exit 0)") {
					t.Errorf("call = %+v", calls[0])
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

func TestDependencyReportHandlerRejectsOversizedBody(t *testing.T) {
	sender := &fakeSender{}
	rec := httptest.NewRecorder()
	req := reportRequest{token: "secret", fields: map[string]string{"exit_code": "1"}, files: map[string]string{"log_tail": strings.Repeat("x", maxReportBody+1)}}.build(t)
	newReportHandler(sender).ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
	if got := sender.snapshot(); len(got) != 0 {
		t.Errorf("calls = %q, want none", kinds(got))
	}
}

func TestDependencyReportHandlerNotMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook/dependency-report", strings.NewReader("exit_code=0"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Report-Token", "secret")
	rec := httptest.NewRecorder()
	newReportHandler(&fakeSender{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestDependencyReportHandlerTelegramFailureDoesNotPanicOrBlock(t *testing.T) {
	sender := &fakeSender{err: errors.New("telegram down")}
	rec := httptest.NewRecorder()
	req := reportRequest{token: "secret", fields: map[string]string{"exit_code": "1"}, files: map[string]string{"log_tail": "x"}}.build(t)
	newReportHandler(sender).ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 even when Telegram fails", rec.Code)
	}
	sender.waitForCalls(t, 1) // the failed send was attempted and only logged
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

func TestDeliverSendsDocumentWithCaption(t *testing.T) {
	sender := &fakeSender{}
	postReport(t, newReportHandlerAt(sender, fixedClock, zap.NewNop()), validReport(t))

	doc := onlyDocument(t, sender.waitForCalls(t, 1))
	if doc.filename != "dependency-report-2026-10-12.html" || doc.chatID != "555" {
		t.Errorf("document = %q to chat %q", doc.filename, doc.chatID)
	}
	if doc.text != fixtureCaption {
		t.Errorf("caption =\n%s\nwant\n%s", doc.text, fixtureCaption)
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
			sender := &fakeSender{}
			postReport(t, newReportHandlerAt(sender, fixedClock, zap.NewNop()), tt.req)
			sender.waitForCalls(t, 1)
			time.Sleep(20 * time.Millisecond) // give a wrong document send a chance to show up
			if got := kinds(sender.snapshot()); !reflect.DeepEqual(got, []string{"message"}) {
				t.Errorf("calls = %q, want one failure message and no document", got)
			}
		})
	}
}

func TestDocumentWhenAllUpToDate(t *testing.T) {
	sender := &fakeSender{}
	req := reportRequest{token: "secret", fields: map[string]string{"exit_code": "0", "pod": "p"}, files: map[string]string{"report": `{"repositories":{}}`}}
	postReport(t, newReportHandlerAt(sender, fixedClock, zap.NewNop()), req)

	doc := onlyDocument(t, sender.waitForCalls(t, 1))
	if doc.filename != "dependency-report-2026-10-12.html" {
		t.Errorf("document = %q", doc.filename)
	}
	want := "📦 <b>Dependency report</b> · 2026-10-12\n\n✅ Todo al día\n\n📊 0 desactualizadas · 0 al día\n🐳 Imágenes: 0 · 📚 Paquetes: 0"
	if doc.text != want {
		t.Errorf("caption = %q, want %q", doc.text, want)
	}
}

func TestDocumentFailureFallsBackToCaptionText(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	sender := &fakeSender{docErr: errors.New("telegram rejected the document")}
	postReport(t, newReportHandlerAt(sender, fixedClock, zap.New(core)), validReport(t))

	calls := sender.waitForCalls(t, 2)
	if got := kinds(calls); !reflect.DeepEqual(got, []string{"document", "message"}) {
		t.Fatalf("calls = %q, want the document, then the caption as text", got)
	}
	if calls[1].text != calls[0].text || !strings.HasPrefix(calls[1].text, "📦 <b>Dependency report</b> ·") {
		t.Errorf("text fallback = %q, want the caption %q", calls[1].text, calls[0].text)
	}
	if n := logs.FilterMessage("Failed to send dependency report document").Len(); n != 1 {
		t.Errorf("document failure logged %d times, want once", n)
	}
}
