package handlers

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
)

// blockingRestarter waits for the rollout until release is closed or its context ends.
type blockingRestarter struct {
	release chan struct{}
}

func (b *blockingRestarter) RestartDeployment(context.Context, string, string) error { return nil }

func (b *blockingRestarter) WaitForRollout(ctx context.Context, _, _ string, _ time.Duration) error {
	select {
	case <-b.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestAllowList(t *testing.T) {
	tests := []struct {
		name    string
		list    allowList
		chatID  string
		allowed bool
	}{
		{name: "empty list denies", list: nil, chatID: "123"},
		{name: "listed chat is allowed", list: allowList{"123", "456"}, chatID: "456", allowed: true},
		{name: "unlisted chat is denied", list: allowList{"123"}, chatID: "999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.list.Authorize(tt.chatID); got != tt.allowed {
				t.Errorf("Authorize(%q) = %v, want %v", tt.chatID, got, tt.allowed)
			}
		})
	}
}

// TestBotTextDefaults checks each bot key renders the text the code sent before the catalog, and
// that an override replaces only its own key.
func TestBotTextDefaults(t *testing.T) {
	cat, err := messages.Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		key  string
		vars map[string]string
		want string
	}{
		{key: "bot.unknown_command", want: "❓ Unknown command."},
		{key: "bot.unauthorized", want: "🚫 You are not authorized to run this command."},
		{
			key:  "bot.workflow_dispatched",
			vars: map[string]string{"workflow": "ops-github-runner-check-updates.yml", "run_url": "https://github.com/devidence-dev/workflows/actions/runs/1"},
			want: "🚀 Workflow *ops-github-runner-check-updates.yml* dispatched successfully. View it at https://github.com/devidence-dev/workflows/actions/runs/1",
		},
		{key: "bot.dispatch_failed", want: "❌ Failed to dispatch the workflow. Check the hub logs or try again later."},
		{key: "bot.scheduled_dispatch_failed", vars: map[string]string{"command": "run_cleanup"}, want: "❌ Scheduled run of `run_cleanup` failed to dispatch. Check the hub logs or try again later."},
		{key: "bot.job_created", vars: map[string]string{"job": "dependency-report-manual-1791763200", "cronjob": "dependency-report"}, want: "🚀 Job *dependency-report-manual-1791763200* created from *dependency-report*."},
		{key: "bot.job_failed", want: "❌ Failed to create the job. Check the hub logs or try again later."},
		{key: "bot.scheduled_job_failed", vars: map[string]string{"command": "run_dependency_report"}, want: "❌ Scheduled run of `run_dependency_report` failed to create the job. Check the hub logs or try again later."},
		{key: "bot.restart_triggered", vars: map[string]string{"deployment": "github-runner"}, want: "🔄 Deployment *github-runner* restart triggered. Waiting for it to come back online..."},
		{key: "bot.restart_failed", want: "❌ Failed to restart the deployment. Check the hub logs or try again later."},
		{key: "bot.restart_ready", vars: map[string]string{"deployment": "github-runner"}, want: "✅ Deployment *github-runner* is up and running again."},
		{key: "bot.restart_timeout", vars: map[string]string{"deployment": "github-runner", "timeout": "1m0s"}, want: "⏱️ Deployment *github-runner* did not report ready within 1m0s — check it manually."},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := cat.Render(tt.key, tt.vars)
			if err != nil || got != tt.want {
				t.Errorf("Render(%s) = %q, %v; want %q", tt.key, got, err, tt.want)
			}
		})
	}

	over, err := messages.Load(map[string]any{"bot": map[string]any{"unknown_command": "❔ Opción inválida."}})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := over.Render("bot.unknown_command", nil); got != "❔ Opción inválida." {
		t.Errorf("override = %q, want the overridden text", got)
	}
	if got, _ := over.Render("bot.unauthorized", nil); got != "🚫 You are not authorized to run this command." {
		t.Errorf("unauthorized = %q, want the default text", got)
	}
}

func TestWorkflowCommandRun(t *testing.T) {
	workflow := config.WorkflowCommandConfig{
		Command: "run_check_updates", Owner: "devidence-dev", Repo: "workflows",
		WorkflowFile: "ops-github-runner-check-updates.yml", Ref: "main", AllowedChatIDs: []string{"123"},
	}
	tests := []struct {
		name        string
		runURL      string
		dispatchErr error
		webURL      string
		want        string
	}{
		{
			name:   "reply carries the run URL",
			runURL: "https://github.com/devidence-dev/workflows/actions/runs/1",
			want:   "🚀 Workflow *ops-github-runner-check-updates.yml* dispatched successfully. View it at https://github.com/devidence-dev/workflows/actions/runs/1",
		},
		{
			name:   "fallback uses the configured web URL",
			webURL: "https://github.example.com",
			want:   "🚀 Workflow *ops-github-runner-check-updates.yml* dispatched successfully. View it at https://github.example.com/devidence-dev/workflows/actions",
		},
		{name: "dispatch error sends the failure text", dispatchErr: errors.New("status 500"), want: "❌ Failed to dispatch the workflow. Check the hub logs or try again later."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := &fakeDispatcher{runURL: tt.runURL, err: tt.dispatchErr}
			messenger := &fakeMessenger{}
			specs := NewCommandSpecs(BotHandlerDeps{GitHubClient: dispatcher, TelegramClient: messenger, Workflows: []config.WorkflowCommandConfig{workflow}, GitHubWebURL: tt.webURL, Logger: zap.NewNop()})

			specs[0].Run(context.Background(), "123")

			if got := messenger.snapshot(); len(got) != 1 || got[0] != tt.want {
				t.Errorf("messages = %#v, want %q", got, tt.want)
			}
		})
	}
}

func TestJobCommandRun(t *testing.T) {
	job := config.JobCommandConfig{Command: "run_dependency_report", Namespace: "internal-apps", CronJob: "dependency-report", AllowedChatIDs: []string{"123"}}
	tests := []struct {
		name      string
		createErr error
		want      string
	}{
		{name: "success replies with the job name", want: "🚀 Job *dependency-report-manual-1791763200* created from *dependency-report*."},
		{name: "API error replies with the failure text", createErr: errors.New("create job: status 403 Forbidden"), want: "❌ Failed to create the job. Check the hub logs or try again later."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator, messenger := &fakeJobCreator{err: tt.createErr}, &fakeMessenger{}
			NewCommandSpecs(BotHandlerDeps{JobClient: creator, TelegramClient: messenger, Jobs: []config.JobCommandConfig{job}, Logger: zap.NewNop()})[0].Run(context.Background(), "123")
			if got := messenger.snapshot(); len(got) != 1 || got[0] != tt.want {
				t.Errorf("messages = %#v, want %q", got, tt.want)
			}
		})
	}
}

func TestJobRunScheduled(t *testing.T) {
	job := config.JobCommandConfig{Command: "run_dependency_report", Namespace: "internal-apps", CronJob: "dependency-report", AllowedChatIDs: []string{"123"}}
	if _, err := NewCommandSpecs(BotHandlerDeps{JobClient: &fakeJobCreator{}, Jobs: []config.JobCommandConfig{job}})[0].RunScheduled(context.Background()); err != nil {
		t.Errorf("RunScheduled() error = %v, want nil", err)
	}
	creator := &fakeJobCreator{err: errors.New("create job: status 403 Forbidden")}
	if _, err := NewCommandSpecs(BotHandlerDeps{JobClient: creator, Jobs: []config.JobCommandConfig{job}})[0].RunScheduled(context.Background()); err == nil {
		t.Error("RunScheduled() error = nil, want the creation error")
	}
}

func TestRestartCommand(t *testing.T) {
	restart := config.RestartCommandConfig{Command: "restart_runner", Namespace: "infrastructure", Deployment: "github-runner", AllowedChatIDs: []string{"123"}}
	tests := []struct {
		name         string
		restartErr   error
		rolloutErr   error
		ready        config.RestartReadyConfig
		wantMessages []string
		wantRollouts int32
	}{
		{name: "triggered and ready", ready: config.RestartReadyConfig{Timeout: time.Minute, PollInterval: time.Second}, wantMessages: []string{"restart triggered", "up and running again"}, wantRollouts: 1},
		{name: "403 sends the failure and starts no wait", restartErr: errors.New("status 403"), wantMessages: []string{"Failed to restart the deployment"}},
		{name: "timeout uses the configured wait", rolloutErr: context.DeadlineExceeded, ready: config.RestartReadyConfig{Timeout: time.Minute, PollInterval: time.Second}, wantMessages: []string{"restart triggered", "within 1m0s"}, wantRollouts: 1},
		{name: "timeout defaults to 3m", rolloutErr: context.DeadlineExceeded, wantMessages: []string{"restart triggered", "within 3m0s"}, wantRollouts: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restarter := &fakeRestarter{err: tt.restartErr, rolloutErr: tt.rolloutErr}
			messenger := &fakeMessenger{}
			specs := NewCommandSpecs(BotHandlerDeps{KubeClient: restarter, TelegramClient: messenger, Restarts: []config.RestartCommandConfig{restart}, RestartReady: tt.ready, Logger: zap.NewNop()})

			specs[0].Run(context.Background(), "123")

			got := waitForMessageCount(t, messenger, len(tt.wantMessages), time.Second)
			for i, want := range tt.wantMessages {
				if !strings.Contains(got[i], want) {
					t.Errorf("message %d = %q, want it to contain %q", i, got[i], want)
				}
			}
			if got := atomic.LoadInt32(&restarter.rolloutCalls); got != tt.wantRollouts {
				t.Errorf("rollout waits = %d, want %d", got, tt.wantRollouts)
			}
		})
	}
}

// TestRestartRunDoesNotWaitForRollout checks that a rollout which never finishes does not hold the
// update loop: Run returns and the next command runs.
func TestRestartRunDoesNotWaitForRollout(t *testing.T) {
	restart := config.RestartCommandConfig{Command: "restart_runner", Namespace: "infrastructure", Deployment: "github-runner", AllowedChatIDs: []string{"123"}}
	restarter := &blockingRestarter{release: make(chan struct{})}
	defer close(restarter.release)
	specs := NewCommandSpecs(BotHandlerDeps{KubeClient: restarter, TelegramClient: &fakeMessenger{}, Restarts: []config.RestartCommandConfig{restart}, RestartReady: config.RestartReadyConfig{Timeout: time.Minute, PollInterval: time.Second}, Logger: zap.NewNop()})

	done := make(chan struct{})
	go func() {
		specs[0].Run(context.Background(), "123")
		specs[0].Run(context.Background(), "123")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run waited for the rollout")
	}
}

func TestScheduledRunnerNotifiesEveryChat(t *testing.T) {
	workflow := config.WorkflowCommandConfig{Command: "run_cleanup", Owner: "devidence-dev", Repo: "workflows", WorkflowFile: "cleanup.yml", AllowedChatIDs: []string{"123", "456"}}

	t.Run("failure sends one message per allowed chat", func(t *testing.T) {
		messenger := &fakeMessenger{}
		deps := BotHandlerDeps{GitHubClient: &fakeDispatcher{err: errors.New("status 500")}, TelegramClient: messenger, Workflows: []config.WorkflowCommandConfig{workflow}, Logger: zap.NewNop()}
		NewScheduledRunner(deps).Wrap(NewCommandSpecs(deps)[0])()
		assertScheduledNotices(t, messenger, workflow, 2)
	})

	t.Run("success logs the dispatched result", func(t *testing.T) {
		core, logs := observer.New(zap.InfoLevel)
		messenger := &fakeMessenger{}
		deps := BotHandlerDeps{GitHubClient: &fakeDispatcher{runURL: "https://github.com/devidence-dev/workflows/actions/runs/1"}, TelegramClient: messenger, Workflows: []config.WorkflowCommandConfig{workflow}, Logger: zap.New(core)}
		NewScheduledRunner(deps).Wrap(NewCommandSpecs(deps)[0])()
		if got := messenger.snapshot(); len(got) != 0 {
			t.Errorf("messages = %#v, want none on success", got)
		}
		entries := logs.FilterMessage("Scheduled command finished").All()
		if len(entries) != 1 || entries[0].ContextMap()["result"] != "dispatched" {
			t.Errorf("log entries = %v, want one with result=dispatched", entries)
		}
	})
}

// assertScheduledNotices checks that wantMessages failure notices reached the workflow's chats in order.
func assertScheduledNotices(t *testing.T, messenger *fakeMessenger, workflow config.WorkflowCommandConfig, wantMessages int) {
	t.Helper()
	if len(messenger.messages) != wantMessages {
		t.Fatalf("messages = %#v, want %d", messenger.messages, wantMessages)
	}
	for i, message := range messenger.messages {
		if !strings.Contains(message, workflow.Command) {
			t.Errorf("message %q does not name the command", message)
		}
		if messenger.chatIDs[i] != workflow.AllowedChatIDs[i] {
			t.Errorf("message %d sent to chat %q, want %q", i, messenger.chatIDs[i], workflow.AllowedChatIDs[i])
		}
	}
}

func TestHandleUnknownCommand(t *testing.T) {
	dispatcher, messenger := &fakeDispatcher{}, &fakeMessenger{}
	handler := NewBotHandler(BotHandlerDeps{GitHubClient: dispatcher, TelegramClient: messenger, Logger: zap.NewNop()})

	handler.Handle(commandUpdate("/deploy_everything", 123))

	if dispatcher.calls != 0 {
		t.Errorf("dispatch calls = %d, want 0", dispatcher.calls)
	}
	if got := messenger.snapshot(); len(got) != 1 || got[0] != "❓ Unknown command." {
		t.Errorf("messages = %#v, want the unknown command text", got)
	}
}

func TestHandleUnauthorizedLogsWarning(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	workflow := config.WorkflowCommandConfig{Command: "run_check_updates", Owner: "devidence-dev", Repo: "workflows", WorkflowFile: "check.yml", AllowedChatIDs: []string{"123"}}
	dispatcher, messenger := &fakeDispatcher{}, &fakeMessenger{}
	handler := NewBotHandler(BotHandlerDeps{GitHubClient: dispatcher, TelegramClient: messenger, Workflows: []config.WorkflowCommandConfig{workflow}, Logger: zap.New(core)})
	update := commandUpdate("/run_check_updates", 999)
	update.Message.From = &tgbotapi.User{UserName: "mallory"}

	handler.Handle(update)

	if dispatcher.calls != 0 {
		t.Errorf("dispatch calls = %d, want 0", dispatcher.calls)
	}
	if got := messenger.snapshot(); len(got) != 1 || got[0] != "🚫 You are not authorized to run this command." {
		t.Errorf("messages = %#v, want the unauthorized text", got)
	}
	warned := logs.FilterMessage("Unauthorized Telegram workflow command attempt").All()
	if len(warned) != 1 {
		t.Fatalf("warnings = %d, want 1", len(warned))
	}
	fields := warned[0].ContextMap()
	if fields["command"] != "run_check_updates" || fields["chat_id"] != "999" || fields["username"] != "mallory" {
		t.Errorf("warning fields = %v, want command, chat_id 999 and username mallory", fields)
	}
}

func TestNewCommandSpecsOrder(t *testing.T) {
	specs := NewCommandSpecs(BotHandlerDeps{
		Workflows: []config.WorkflowCommandConfig{{Command: "run_cleanup"}},
		Restarts:  []config.RestartCommandConfig{{Command: "restart_runner"}},
		Jobs:      []config.JobCommandConfig{{Command: "run_dependency_report"}},
	})
	var got []string
	for _, spec := range specs {
		got = append(got, spec.Name())
	}
	if want := "run_cleanup,restart_runner,run_dependency_report"; strings.Join(got, ",") != want {
		t.Errorf("specs = %v, want %s", got, want)
	}
}

func TestNewCommandSpecsWorkflowsOnlyLeavesKubeNil(t *testing.T) {
	deps := BotHandlerDeps{GitHubClient: &fakeDispatcher{}, Workflows: []config.WorkflowCommandConfig{{Command: "run_cleanup"}}}
	if specs := NewCommandSpecs(deps); len(specs) != 1 {
		t.Fatalf("specs = %d, want 1", len(specs))
	}
	if d := deps.withDefaults(); d.KubeClient != nil || d.JobClient != nil {
		t.Errorf("KubeClient = %v, JobClient = %v, want nil interfaces", d.KubeClient, d.JobClient)
	}
}
