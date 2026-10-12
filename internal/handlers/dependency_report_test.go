package handlers

import (
	"bytes"
	"errors"
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
	"automation-hub/internal/messages"
	"automation-hub/internal/services/depreport"
)

const (
	testForkRepo = "devidence-dev/discord-tts-bot"
	// testMaxBody is the default body limit of the config (10 MiB).
	testMaxBody = 10 << 20
	testToken   = "secret"
	testChatID  = "555"
)

// fixedClock is the time the hub "receives" the report in the handler tests: 2026-10-12 09:07 Guayaquil.
var fixedClock = time.Date(2026, 10, 12, 14, 7, 0, 0, time.UTC)

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
	header string // token header; empty means X-Report-Token
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
		header := rr.header
		if header == "" {
			header = "X-Report-Token"
		}
		req.Header.Set(header, rr.token)
	}
	return req
}

// testReportConfig is the config of the handler tests: the defaults of config.Load.
func testReportConfig() config.DependencyReportConfig {
	return config.DependencyReportConfig{
		WebhookToken:   testToken,
		TelegramChatID: testChatID,
		Forks:          []string{testForkRepo},
		Path:           "/webhook/dependency-report",
		TokenHeader:    "X-Report-Token",
		MaxBodyBytes:   testMaxBody,
		DocumentName:   "dependency-report-{date}.html",
		DateLayout:     "2006-01-02",
	}
}

// newReportHandlerAt returns the handler with the default texts, the Guayaquil zone and the fixed clock.
func newReportHandlerAt(t *testing.T, sender depreport.ReportSender, now time.Time, logger *zap.Logger) *DependencyReportHandler {
	t.Helper()
	cat, err := messages.Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	loc, err := time.LoadLocation("America/Guayaquil")
	if err != nil {
		t.Fatal(err)
	}
	h := NewDependencyReportHandler(testReportConfig(), cat, loc, sender, logger)
	h.now = func() time.Time { return now }
	return h
}

func newReportHandler(t *testing.T, sender depreport.ReportSender) *DependencyReportHandler {
	t.Helper()
	return newReportHandlerAt(t, sender, fixedClock, zap.NewNop())
}

func fixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/renovate-report.json")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// validReport is a successful request for the fixture.
func validReport(t *testing.T) reportRequest {
	t.Helper()
	return reportRequest{token: testToken, fields: map[string]string{"exit_code": "0", "pod": "p"}, files: map[string]string{"report": fixture(t)}}
}

// postReport posts req and requires 202.
func postReport(t *testing.T, h *DependencyReportHandler, req reportRequest) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.build(t))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 (%s)", rec.Code, rec.Body.String())
	}
}

// reportHandlerCase is one request to the handler and what it must answer and send.
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
	newReportHandler(t, sender).ServeHTTP(rec, tt.req.build(t))
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
		if c.chatID != testChatID {
			t.Errorf("chat = %q, want the configured one", c.chatID)
		}
	}
	if tt.check != nil {
		tt.check(t, calls)
	}
}

func TestDependencyReportHandler(t *testing.T) {
	fix := fixture(t)
	tests := []reportHandlerCase{
		{name: "missing token", req: reportRequest{fields: map[string]string{"exit_code": "0"}, files: map[string]string{"report": fix}}, wantStatus: http.StatusUnauthorized},
		{name: "wrong token", req: reportRequest{token: "nope", fields: map[string]string{"exit_code": "0"}, files: map[string]string{"report": fix}}, wantStatus: http.StatusUnauthorized},
		{name: "non-integer exit code", req: reportRequest{token: testToken, fields: map[string]string{"exit_code": "abc"}}, wantStatus: http.StatusBadRequest},
		{name: "missing exit code", req: reportRequest{token: testToken, fields: map[string]string{"pod": "p"}}, wantStatus: http.StatusBadRequest},
		{name: "invalid report json", req: reportRequest{token: testToken, fields: map[string]string{"exit_code": "0"}, files: map[string]string{"report": "{not json"}}, wantStatus: http.StatusBadRequest},
		{
			name: "valid report; forks come from config, not the request",
			req: reportRequest{
				token:  testToken,
				fields: map[string]string{"exit_code": "0", "pod": "p", "forks": `["devidence-dev/QuantWarden"]`},
				files:  map[string]string{"report": fix, "log_tail": "ignored"},
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
			req:        reportRequest{token: testToken, fields: map[string]string{"exit_code": "1", "pod": "pod-x"}, files: map[string]string{"log_tail": "<html>"}},
			wantStatus: http.StatusAccepted, wantCalls: 1,
			check: func(t *testing.T, calls []sentCall) {
				if calls[0].kind != "message" || !strings.HasPrefix(calls[0].text, "❌ Renovate falló (exit 1) en el pod pod-x") || !strings.Contains(calls[0].text, "&lt;html&gt;") {
					t.Errorf("call = %+v", calls[0])
				}
			},
		},
		{
			name:       "exit 0 without report is a failure",
			req:        reportRequest{token: testToken, fields: map[string]string{"exit_code": "0", "pod": "pod-x"}},
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

func TestAuthRejectsMissingToken(t *testing.T) {
	sender := &fakeSender{}
	rec := httptest.NewRecorder()
	req := reportRequest{fields: map[string]string{"exit_code": "1"}, files: map[string]string{"log_tail": "x"}}.build(t)
	newReportHandler(t, sender).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	time.Sleep(20 * time.Millisecond)
	if got := sender.snapshot(); len(got) != 0 {
		t.Errorf("calls = %q, want none", kinds(got))
	}
}

func TestCustomTokenHeader(t *testing.T) {
	cfg := testReportConfig()
	cfg.TokenHeader = "X-Deps-Token"
	h := newReportHandler(t, &fakeSender{})
	h.cfg = cfg

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, reportRequest{token: testToken, fields: map[string]string{"exit_code": "1"}, files: map[string]string{"log_tail": "x"}}.build(t))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("default header accepted: status = %d, want 401", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, reportRequest{token: testToken, header: "X-Deps-Token", fields: map[string]string{"exit_code": "1"}, files: map[string]string{"log_tail": "x"}}.build(t))
	if rec.Code != http.StatusAccepted {
		t.Errorf("configured header refused: status = %d, want 202", rec.Code)
	}
}

func TestBodyTooLargeGets413(t *testing.T) {
	sender := &fakeSender{}
	rec := httptest.NewRecorder()
	req := reportRequest{token: testToken, fields: map[string]string{"exit_code": "1"}, files: map[string]string{"log_tail": strings.Repeat("x", testMaxBody+1)}}.build(t)
	newReportHandler(t, sender).ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
	if got := sender.snapshot(); len(got) != 0 {
		t.Errorf("calls = %q, want none", kinds(got))
	}
}

func TestMalformedExitCodeGets400(t *testing.T) {
	rec := httptest.NewRecorder()
	req := reportRequest{token: testToken, fields: map[string]string{"exit_code": "abc"}}.build(t)
	newReportHandler(t, &fakeSender{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestNotMultipartGets400(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/webhook/dependency-report", strings.NewReader("exit_code=0"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Report-Token", testToken)
	rec := httptest.NewRecorder()
	newReportHandler(t, &fakeSender{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAcceptsBeforeSending(t *testing.T) {
	gate := make(chan struct{})
	sender := &gatedSender{fakeSender: &fakeSender{}, gate: gate}
	rec := httptest.NewRecorder()
	newReportHandler(t, sender).ServeHTTP(rec, validReport(t).build(t))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", rec.Code)
	}
	if got := sender.snapshot(); len(got) != 0 {
		t.Errorf("sent %q before the 202 was read; want none yet", kinds(got))
	}
	close(gate)
	sender.waitForCalls(t, 1)
}

// gatedSender holds every send until gate is closed.
type gatedSender struct {
	*fakeSender
	gate chan struct{}
}

func (g *gatedSender) SendDocument(chatID, filename, caption string, data []byte) error {
	<-g.gate
	return g.fakeSender.SendDocument(chatID, filename, caption, data)
}

// countingAggregate wraps depreport.Aggregate and counts its calls.
type countingAggregate struct{ calls int }

func (c *countingAggregate) aggregate(rep depreport.Report, forks []string, receivedAt time.Time) depreport.Summary {
	c.calls++
	return depreport.Aggregate(rep, forks, receivedAt)
}

func TestAggregatesOnce(t *testing.T) {
	h := newReportHandler(t, &fakeSender{})
	counter := &countingAggregate{}
	h.aggregate = counter.aggregate
	postReport(t, h, validReport(t))
	deadline := time.Now().Add(2 * time.Second)
	for counter.calls == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if counter.calls != 1 {
		t.Errorf("aggregate calls = %d, want 1", counter.calls)
	}
}

func TestDeliverSendsDocumentWithCaption(t *testing.T) {
	sender := &fakeSender{}
	postReport(t, newReportHandler(t, sender), validReport(t))

	doc := onlyDocument(t, sender.waitForCalls(t, 1))
	if doc.filename != "dependency-report-2026-10-12.html" || doc.chatID != testChatID {
		t.Errorf("document = %q to chat %q", doc.filename, doc.chatID)
	}
	if want := readGolden(t, "caption.txt"); doc.text != want {
		t.Errorf("caption =\n%s\nwant\n%s", doc.text, want)
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
		{name: "exit code 1 with a report", req: reportRequest{token: testToken, fields: map[string]string{"exit_code": "1", "pod": "p"}, files: map[string]string{"report": "{}", "log_tail": "x"}}},
		{name: "exit code 0 without a report", req: reportRequest{token: testToken, fields: map[string]string{"exit_code": "0", "pod": "p"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeSender{}
			postReport(t, newReportHandler(t, sender), tt.req)
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
	req := reportRequest{token: testToken, fields: map[string]string{"exit_code": "0", "pod": "p"}, files: map[string]string{"report": `{"repositories":{}}`}}
	postReport(t, newReportHandler(t, sender), req)

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
	h := newReportHandlerAt(t, sender, fixedClock, zap.New(core))
	postReport(t, h, validReport(t))

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

func TestDependencyReportHandlerTelegramFailureDoesNotPanicOrBlock(t *testing.T) {
	sender := &fakeSender{err: errors.New("telegram down")}
	rec := httptest.NewRecorder()
	req := reportRequest{token: testToken, fields: map[string]string{"exit_code": "1"}, files: map[string]string{"log_tail": "x"}}.build(t)
	newReportHandler(t, sender).ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202 even when Telegram fails", rec.Code)
	}
	sender.waitForCalls(t, 1) // the failed send was attempted and only logged
}
