package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"automation-hub/internal/config"
	"automation-hub/internal/handlers"
	"automation-hub/internal/messages"
)

type noopSender struct{}

func (noopSender) SendHTML(string, string) error { return nil }

func (noopSender) SendDocument(string, string, string, []byte) error { return nil }

// routesFor registers the routes of cfg on a fresh router, with the default catalog and zone.
func routesFor(t *testing.T, cfg *config.Config) *mux.Router {
	t.Helper()
	cat, err := messages.Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	loc, err := time.LoadLocation("America/Guayaquil")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Catalog, cfg.Location = cat, loc
	router := mux.NewRouter()
	registerRoutes(router, cfg, handlers.NewWebhookHandler(nil, cfg, zap.NewNop()), noopSender{}, zap.NewNop())
	return router
}

// jobPost is a Job request for a failed Renovate run, which needs no report file.
func jobPost(t *testing.T, path, token string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if err := mw.WriteField("exit_code", "1"); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-Report-Token", token)
	return req
}

func TestRegisterRoutesDependencyReport(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		token      string
		wantStatus int
	}{
		{name: "custom path is accepted", path: "/hooks/deps", token: "secret", wantStatus: http.StatusAccepted},
		{name: "custom path needs the token", path: "/hooks/deps", wantStatus: http.StatusUnauthorized},
		{name: "old path is gone", path: "/webhook/dependency-report", token: "secret", wantStatus: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := routesFor(t, &config.Config{DependencyReport: &config.DependencyReportConfig{
				WebhookToken: "secret", TelegramChatID: "1", Path: "/hooks/deps", TokenHeader: "X-Report-Token",
				MaxBodyBytes: 1 << 20, DocumentName: "dependency-report-{date}.html", DateLayout: "2006-01-02",
			}})
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, jobPost(t, tt.path, tt.token))
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestRegisterRoutesNotConfigured(t *testing.T) {
	router := routesFor(t, &config.Config{})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, jobPost(t, "/webhook/dependency-report", "secret"))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 without the dependency_report section", rec.Code)
	}
}

// TestJobSchedulesUseConfiguredZone registers a job schedule in a non-default zone: the scheduler
// runs in that zone, so the next run is Monday 09:00 in Europe/Madrid.
func TestJobSchedulesUseConfiguredZone(t *testing.T) {
	madrid, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		Location: madrid,
		Jobs: []config.JobCommandConfig{{
			Command: "run_dependency_report", Namespace: "internal-apps", CronJob: "dependency-report", Schedule: "0 9 * * 1",
		}},
	}
	core, logs := observer.New(zap.InfoLevel)
	sched := registerSchedules(cfg, handlers.NewBotHandler(handlers.BotHandlerDeps{}), cfg.Location, zap.New(core))
	if sched == nil {
		t.Fatal("registerSchedules() returned no scheduler for a scheduled job")
	}
	sched.Start()
	defer sched.Stop()

	started := logs.FilterMessage("Scheduler started").All()
	if len(started) != 1 || started[0].ContextMap()["timezone"] != "Europe/Madrid" {
		t.Fatalf("scheduler started = %v, want timezone Europe/Madrid", started)
	}
	next := logs.FilterMessage("Next scheduled run").All()
	if len(next) != 1 {
		t.Fatalf("next runs logged = %d, want 1", len(next))
	}
	nextRun, _ := next[0].ContextMap()["next_run"].(string)
	at, err := time.Parse(time.RFC3339, nextRun)
	if err != nil {
		t.Fatalf("next_run %q: %v", nextRun, err)
	}
	if local := at.In(madrid); local.Weekday() != time.Monday || local.Hour() != 9 || local.Minute() != 0 {
		t.Errorf("next run = %s, want Monday 09:00 in Europe/Madrid", local)
	}
}
