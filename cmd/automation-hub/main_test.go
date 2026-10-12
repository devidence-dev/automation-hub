package main

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	return routesWith(t, cfg, zap.NewNop())
}

// routesWith is routesFor with the given logger, so a test can read what the registration logged.
func routesWith(t *testing.T, cfg *config.Config, logger *zap.Logger) *mux.Router {
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
	registerRoutes(router, cfg, handlers.RouteEnv{
		DependencyReport: cfg.DependencyReport,
		Catalog:          cat,
		Location:         loc,
		Sender:           noopSender{},
		Logger:           logger,
	})
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

func TestRegisterRoutesQbittorrent(t *testing.T) {
	cfg := &config.Config{Hook: []config.WebhookConfig{{
		Name: "qbittorrent", Path: "/webhook/qbittorrent",
		Config: config.WebhookProcessorConfig{TelegramChatID: "123", TelegramMessage: "Downloaded: %s at %s"},
	}}}
	router := routesFor(t, cfg)
	post := httptest.NewRecorder()
	router.ServeHTTP(post, httptest.NewRequest(http.MethodPost, "/webhook/qbittorrent", strings.NewReader(`{"torrent_name":"Debian ISO","save_path":"/downloads"}`)))
	if post.Code != http.StatusOK {
		t.Errorf("POST status = %d, want 200", post.Code)
	}
	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/webhook/qbittorrent", nil))
	if get.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET status = %d, want 405", get.Code)
	}
}

func TestRegisterRoutesUnknownHookWarns(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	cfg := &config.Config{Hook: []config.WebhookConfig{{Name: "unknown-hook", Path: "/webhook/unknown"}}}
	router := routesWith(t, cfg, zap.New(core))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhook/unknown", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for an unknown hook", rec.Code)
	}
	warned := logs.FilterMessage("Unknown webhook type").All()
	if len(warned) != 1 || warned[0].ContextMap()["name"] != "unknown-hook" {
		t.Errorf("warnings = %v, want one naming unknown-hook", warned)
	}
}

// TestRegisterRoutesQbittorrentMissing checks that without a qbittorrent hook its path has no route.
func TestRegisterRoutesQbittorrentMissing(t *testing.T) {
	cfg := &config.Config{Hook: []config.WebhookConfig{{Name: "sonarr", Path: "/webhook/sonarr"}}}
	router := routesWith(t, cfg, zap.NewNop())

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhook/qbittorrent", strings.NewReader("{")))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 without a qbittorrent hook", rec.Code)
	}
}

// TestRegisterRoutesHookNamedDependencyReportIgnored checks that a hook entry named dependency-report
// is unknown: the report route mounts once, from its section, and the entry only logs a warning.
func TestRegisterRoutesHookNamedDependencyReportIgnored(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	cfg := &config.Config{
		Hook: []config.WebhookConfig{{Name: "dependency-report", Path: "/hooks/deps"}},
		DependencyReport: &config.DependencyReportConfig{
			WebhookToken: "secret", TelegramChatID: "1", Path: "/hooks/deps", TokenHeader: "X-Report-Token",
			MaxBodyBytes: 1 << 20, DocumentName: "dependency-report-{date}.html", DateLayout: "2006-01-02",
		},
	}
	router := routesWith(t, cfg, zap.New(core))

	warned := logs.FilterMessage("Unknown webhook type").All()
	if len(warned) != 1 || warned[0].ContextMap()["name"] != "dependency-report" {
		t.Errorf("warnings = %v, want one naming dependency-report", warned)
	}
	if registered := logs.FilterMessage("Registered webhook route").All(); len(registered) != 1 {
		t.Errorf("registered routes = %d, want 1", len(registered))
	}
	mounted := 0
	err := router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		tmpl, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		if tmpl == "/hooks/deps" {
			mounted++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if mounted != 1 {
		t.Errorf("routes on /hooks/deps = %d, want 1", mounted)
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
	deps := handlers.BotHandlerDeps{Jobs: cfg.Jobs, Logger: zap.New(core)}
	lazy := &lazyScheduler{loc: cfg.Location, logger: zap.New(core)}
	if err := registerCommands(&fakeMenu{}, lazy, handlers.NewScheduledRunner(deps), handlers.NewCommandSpecs(deps)); err != nil {
		t.Fatalf("registerCommands() error = %v", err)
	}
	sched := lazy.sched
	if sched == nil {
		t.Fatal("registerCommands() created no scheduler for a scheduled job")
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

// fakeMenu records the commands published to Telegram, and returns err when it is set.
type fakeMenu struct {
	commands []tgbotapi.BotCommand
	err      error
}

func (f *fakeMenu) SetMyCommands(commands []tgbotapi.BotCommand) error {
	f.commands = commands
	return f.err
}

// fakeAdder records the names of the schedules registered.
type fakeAdder struct {
	names []string
}

func (f *fakeAdder) Add(name, _ string, _ func()) error {
	f.names = append(f.names, name)
	return nil
}

func TestRegisterCommandsOneLoop(t *testing.T) {
	deps := handlers.BotHandlerDeps{
		Workflows: []config.WorkflowCommandConfig{{Command: "run_cleanup", Description: "Cleanup", Schedule: "0 9 * * 2,5"}},
		Restarts:  []config.RestartCommandConfig{{Command: "restart_runner", Description: "Restart"}},
		Jobs:      []config.JobCommandConfig{{Command: "run_dependency_report", Description: "Report", Schedule: "0 9 * * 1"}},
	}
	menu, adder := &fakeMenu{}, &fakeAdder{}
	if err := registerCommands(menu, adder, handlers.NewScheduledRunner(deps), handlers.NewCommandSpecs(deps)); err != nil {
		t.Fatalf("registerCommands() error = %v", err)
	}

	var got []string
	for _, c := range menu.commands {
		got = append(got, c.Command)
	}
	if want := []string{"run_cleanup", "restart_runner", "run_dependency_report"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("menu = %v, want %v", got, want)
	}
	if want := []string{"run_cleanup", "run_dependency_report"}; strings.Join(adder.names, ",") != strings.Join(want, ",") {
		t.Errorf("schedules = %v, want %v: the restart has no schedule", adder.names, want)
	}
}

// TestRegisterCommandsMenuRejectedReturnsError checks that a menu Telegram rejects comes back as an
// error, so main exits before it starts polling.
func TestRegisterCommandsMenuRejectedReturnsError(t *testing.T) {
	rejected := errors.New("telegram: 400 Bad Request")
	deps := handlers.BotHandlerDeps{Workflows: []config.WorkflowCommandConfig{{Command: "run_cleanup", Description: "Cleanup"}}}
	err := registerCommands(&fakeMenu{err: rejected}, &fakeAdder{}, handlers.NewScheduledRunner(deps), handlers.NewCommandSpecs(deps))
	if !errors.Is(err, rejected) {
		t.Errorf("registerCommands() error = %v, want it to wrap %v", err, rejected)
	}
}

// TestNewHTTPServerUsesConfiguredTimeouts checks that the server takes its read and write timeouts from the config.
func TestNewHTTPServerUsesConfiguredTimeouts(t *testing.T) {
	cfg := &config.Config{Server: config.ServerConfig{
		Address:         ":0",
		ReadTimeout:     7 * time.Second,
		WriteTimeout:    9 * time.Second,
		ShutdownTimeout: 11 * time.Second,
	}}
	srv := newHTTPServer(cfg, nil, zap.NewNop())
	if srv.ReadTimeout != 7*time.Second || srv.WriteTimeout != 9*time.Second {
		t.Errorf("timeouts = read %s, write %s; want 7s and 9s", srv.ReadTimeout, srv.WriteTimeout)
	}
}

func TestNewGitHubClientNilWhenUnconfigured(t *testing.T) {
	if client := newGitHubClient(&config.Config{}, zap.NewNop()); client != nil {
		t.Errorf("newGitHubClient() = %#v, want a nil interface without workflows", client)
	}
}

func TestNewKubeClientNilWhenUnconfigured(t *testing.T) {
	if client := newKubeClient(&config.Config{}, zap.NewNop()); client != nil {
		t.Errorf("newKubeClient() = %#v, want a nil interface without restarts or jobs", client)
	}
}
