package handlers

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"automation-hub/internal/config"
)

// fakeDispatcher returns runURL on success; an empty runURL models a 204 answer.
type fakeDispatcher struct {
	calls  int
	err    error
	runURL string
}

func (f *fakeDispatcher) DispatchWorkflow(_ context.Context, owner, repo, workflowFile, ref string) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.runURL, nil
}

type fakeRestarter struct {
	calls        int
	err          error
	rolloutCalls int32
	rolloutErr   error
}

func (f *fakeRestarter) RestartDeployment(_ context.Context, namespace, deployment string) error {
	f.calls++
	return f.err
}

func (f *fakeRestarter) WaitForRollout(_ context.Context, namespace, deployment string, pollInterval time.Duration) error {
	atomic.AddInt32(&f.rolloutCalls, 1)
	return f.rolloutErr
}

type fakeJobCreator struct {
	mu         sync.Mutex
	namespaces []string
	cronJobs   []string
	err        error
}

func (f *fakeJobCreator) CreateJobFromCronJob(_ context.Context, namespace, name string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.namespaces = append(f.namespaces, namespace)
	f.cronJobs = append(f.cronJobs, name)
	if f.err != nil {
		return "", f.err
	}
	return name + "-manual-1791763200", nil
}

type fakeMessenger struct {
	mu       sync.Mutex
	chatIDs  []string
	messages []string
}

func (f *fakeMessenger) SendMessage(chatID, message string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.chatIDs = append(f.chatIDs, chatID)
	f.messages = append(f.messages, message)
	return nil
}

func (f *fakeMessenger) snapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.messages...)
}

// waitForMessageCount polls until the messenger has at least n messages, since
// restartDeployment's rollout confirmation is sent from a background goroutine.
func waitForMessageCount(t *testing.T, messenger *fakeMessenger, n int, timeout time.Duration) []string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		messages := messenger.snapshot()
		if len(messages) >= n {
			return messages
		}
		if time.Now().After(deadline) {
			t.Fatalf("messages = %#v after %s, want at least %d", messages, timeout, n)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestBotHandlerHandle(t *testing.T) {
	workflow := config.WorkflowCommandConfig{
		Command: "run_check_updates", Owner: "devidence-dev", Repo: "github-runner",
		WorkflowFile: "check-updates.yml", Ref: "master", AllowedChatIDs: []string{"123"},
	}
	tests := []struct {
		name          string
		update        tgbotapi.Update
		dispatchError error
		wantCalls     int
		wantMessage   string
	}{
		{name: "authorized command dispatches workflow", update: commandUpdate("/run_check_updates", 123), wantCalls: 1, wantMessage: "dispatched successfully"},
		{name: "unauthorized chat is rejected", update: commandUpdate("/run_check_updates", 999), wantMessage: "not authorized"},
		{name: "unknown command does not dispatch", update: commandUpdate("/run_deploy_prod", 123), wantMessage: "Unknown command"},
		{name: "github failure is reported", update: commandUpdate("/run_check_updates", 123), dispatchError: context.DeadlineExceeded, wantCalls: 1, wantMessage: "Failed to dispatch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := &fakeDispatcher{err: tt.dispatchError, runURL: "https://github.com/devidence-dev/github-runner/actions/runs/999"}
			messenger := &fakeMessenger{}
			handler := NewBotHandler(BotHandlerDeps{GitHubClient: dispatcher, TelegramClient: messenger, Workflows: []config.WorkflowCommandConfig{workflow}, Logger: zap.NewNop()})
			handler.Handle(tt.update)
			if dispatcher.calls != tt.wantCalls {
				t.Errorf("dispatch calls = %d, want %d", dispatcher.calls, tt.wantCalls)
			}
			if tt.wantMessage != "" {
				if len(messenger.messages) != 1 || !strings.Contains(messenger.messages[0], tt.wantMessage) {
					t.Errorf("messages = %#v, want one containing %q", messenger.messages, tt.wantMessage)
				}
			}
		})
	}
}

func TestBotHandlerDispatchWorkflowLinksToTheDispatchedRun(t *testing.T) {
	workflow := config.WorkflowCommandConfig{
		Command: "run_check_updates", Owner: "devidence-dev", Repo: "github-runner",
		WorkflowFile: "check-updates.yml", Ref: "main", AllowedChatIDs: []string{"123"},
	}
	dispatcher := &fakeDispatcher{runURL: "https://github.com/devidence-dev/github-runner/actions/runs/999"}
	messenger := &fakeMessenger{}
	handler := NewBotHandler(BotHandlerDeps{GitHubClient: dispatcher, TelegramClient: messenger, Workflows: []config.WorkflowCommandConfig{workflow}, Logger: zap.NewNop()})

	handler.Handle(commandUpdate("/run_check_updates", 123))

	if len(messenger.messages) != 1 || !strings.Contains(messenger.messages[0], dispatcher.runURL) {
		t.Fatalf("messages = %#v, want one containing the dispatched run URL %q", messenger.messages, dispatcher.runURL)
	}
	if strings.Contains(messenger.messages[0], "/actions\n") || strings.HasSuffix(messenger.messages[0], "/actions") {
		t.Errorf("message links to the general Actions tab instead of the dispatched run: %q", messenger.messages[0])
	}
}

func TestBotHandlerDispatchWorkflowFallsBackWhenRunURLIsEmpty(t *testing.T) {
	workflow := config.WorkflowCommandConfig{
		Command: "run_check_updates", Owner: "devidence-dev", Repo: "github-runner",
		WorkflowFile: "check-updates.yml", Ref: "main", AllowedChatIDs: []string{"123"},
	}
	dispatcher := &fakeDispatcher{runURL: ""}
	messenger := &fakeMessenger{}
	handler := NewBotHandler(BotHandlerDeps{GitHubClient: dispatcher, TelegramClient: messenger, Workflows: []config.WorkflowCommandConfig{workflow}, Logger: zap.NewNop()})

	handler.Handle(commandUpdate("/run_check_updates", 123))

	wantFallback := "https://github.com/devidence-dev/github-runner/actions"
	if len(messenger.messages) != 1 || !strings.Contains(messenger.messages[0], wantFallback) {
		t.Fatalf("messages = %#v, want one containing the fallback URL %q", messenger.messages, wantFallback)
	}
}

func TestBotHandlerHandleRestart(t *testing.T) {
	restart := config.RestartCommandConfig{
		Command: "restart_runner", Namespace: "infrastructure", Deployment: "github-runner", AllowedChatIDs: []string{"123"},
	}
	tests := []struct {
		name        string
		update      tgbotapi.Update
		restartErr  error
		wantCalls   int
		wantMessage string
	}{
		{name: "authorized command triggers restart", update: commandUpdate("/restart_runner", 123), wantCalls: 1, wantMessage: "restart triggered"},
		{name: "unauthorized chat is rejected", update: commandUpdate("/restart_runner", 999), wantMessage: "not authorized"},
		{name: "restart failure is reported", update: commandUpdate("/restart_runner", 123), restartErr: context.DeadlineExceeded, wantCalls: 1, wantMessage: "Failed to restart"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restarter := &fakeRestarter{err: tt.restartErr}
			messenger := &fakeMessenger{}
			handler := NewBotHandler(BotHandlerDeps{KubeClient: restarter, TelegramClient: messenger, Restarts: []config.RestartCommandConfig{restart}, Logger: zap.NewNop()})
			handler.Handle(tt.update)
			if restarter.calls != tt.wantCalls {
				t.Errorf("restart calls = %d, want %d", restarter.calls, tt.wantCalls)
			}
			if tt.wantMessage != "" {
				messages := messenger.snapshot()
				if len(messages) != 1 || !strings.Contains(messages[0], tt.wantMessage) {
					t.Errorf("messages = %#v, want one containing %q", messages, tt.wantMessage)
				}
			}
		})
	}
}

func TestBotHandlerRestartConfirmsWhenRolloutBecomesReady(t *testing.T) {
	restart := config.RestartCommandConfig{
		Command: "restart_runner", Namespace: "infrastructure", Deployment: "github-runner", AllowedChatIDs: []string{"123"},
	}
	restarter := &fakeRestarter{}
	messenger := &fakeMessenger{}
	handler := NewBotHandler(BotHandlerDeps{KubeClient: restarter, TelegramClient: messenger, Restarts: []config.RestartCommandConfig{restart}, Logger: zap.NewNop()})

	handler.Handle(commandUpdate("/restart_runner", 123))

	messages := waitForMessageCount(t, messenger, 2, time.Second)
	if !strings.Contains(messages[0], "Waiting for it to come back online") {
		t.Errorf("first message = %q, want the immediate trigger confirmation", messages[0])
	}
	if !strings.Contains(messages[1], "up and running again") {
		t.Errorf("second message = %q, want the rollout-ready confirmation", messages[1])
	}
}

func TestBotHandlerRestartReportsTimeoutWhenRolloutNeverBecomesReady(t *testing.T) {
	restart := config.RestartCommandConfig{
		Command: "restart_runner", Namespace: "infrastructure", Deployment: "github-runner", AllowedChatIDs: []string{"123"},
	}
	restarter := &fakeRestarter{rolloutErr: context.DeadlineExceeded}
	messenger := &fakeMessenger{}
	handler := NewBotHandler(BotHandlerDeps{KubeClient: restarter, TelegramClient: messenger, Restarts: []config.RestartCommandConfig{restart}, Logger: zap.NewNop()})

	handler.Handle(commandUpdate("/restart_runner", 123))

	messages := waitForMessageCount(t, messenger, 2, time.Second)
	if !strings.Contains(messages[1], "did not report ready") {
		t.Errorf("second message = %q, want the timeout notice", messages[1])
	}
}

func commandUpdate(command string, chatID int64) tgbotapi.Update {
	return tgbotapi.Update{Message: &tgbotapi.Message{
		Text:     command,
		Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(command)}},
		Chat:     &tgbotapi.Chat{ID: chatID},
	}}
}

func TestRunWorkflow(t *testing.T) {
	workflow := config.WorkflowCommandConfig{
		Command: "run_check_updates", Owner: "devidence-dev", Repo: "workflows",
		WorkflowFile: "ops-github-runner-check-updates.yml", Ref: "main",
	}
	tests := []struct {
		name        string
		runURL      string
		dispatchErr error
		want        string
		wantErr     bool
	}{
		{name: "returns the html_url of the created run", runURL: "https://github.com/devidence-dev/workflows/actions/runs/123456", want: "https://github.com/devidence-dev/workflows/actions/runs/123456"},
		{name: "falls back to the Actions page without run details", want: "https://github.com/devidence-dev/workflows/actions"},
		{name: "dispatch error is returned", dispatchErr: errors.New("status 500"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := &fakeDispatcher{runURL: tt.runURL, err: tt.dispatchErr}
			handler := NewBotHandler(BotHandlerDeps{GitHubClient: dispatcher, Logger: zap.NewNop()})

			got, err := handler.RunWorkflow(context.Background(), workflow)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RunWorkflow() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("RunWorkflow() = %q, want %q", got, tt.want)
			}
			if dispatcher.calls != 1 {
				t.Errorf("dispatch calls = %d, want 1", dispatcher.calls)
			}
		})
	}
}

func TestRunScheduled(t *testing.T) {
	workflow := config.WorkflowCommandConfig{
		Command: "run_cleanup", Owner: "devidence-dev", Repo: "workflows",
		WorkflowFile: "ops-github-runner-cleanup.yml", Ref: "main", AllowedChatIDs: []string{"123", "456"},
	}
	tests := []struct {
		name         string
		dispatchErr  error
		wantMessages int
	}{
		{name: "success sends no chat message"},
		{name: "failure notifies every allowed chat", dispatchErr: errors.New("status 500"), wantMessages: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := &fakeDispatcher{runURL: "https://github.com/devidence-dev/workflows/actions/runs/1", err: tt.dispatchErr}
			messenger := &fakeMessenger{}
			handler := NewBotHandler(BotHandlerDeps{GitHubClient: dispatcher, TelegramClient: messenger, Workflows: []config.WorkflowCommandConfig{workflow}, Logger: zap.NewNop()})

			handler.RunScheduled(workflow)

			assertScheduledRun(t, dispatcher, messenger, workflow, tt.wantMessages)
		})
	}
}

// assertScheduledRun checks one scheduled run: one dispatch, then wantMessages
// chat messages, each naming the command and sent to its allowed chat in order.
func assertScheduledRun(t *testing.T, dispatcher *fakeDispatcher, messenger *fakeMessenger, workflow config.WorkflowCommandConfig, wantMessages int) {
	t.Helper()
	if dispatcher.calls != 1 {
		t.Errorf("dispatch calls = %d, want 1", dispatcher.calls)
	}
	if len(messenger.messages) != wantMessages {
		t.Fatalf("messages = %#v, want %d", messenger.messages, wantMessages)
	}
	for i, message := range messenger.messages {
		if !strings.Contains(message, "run_cleanup") {
			t.Errorf("message %q does not name the command", message)
		}
		if messenger.chatIDs[i] != workflow.AllowedChatIDs[i] {
			t.Errorf("message %d sent to chat %q, want %q", i, messenger.chatIDs[i], workflow.AllowedChatIDs[i])
		}
	}
}

func TestBotHandlerHandleJob(t *testing.T) {
	job := config.JobCommandConfig{
		Command: "run_dependency_report", Namespace: "internal-apps", CronJob: "dependency-report", AllowedChatIDs: []string{"123"},
	}
	tests := []struct {
		name        string
		update      tgbotapi.Update
		createErr   error
		wantCalls   int
		wantMessage string
	}{
		{name: "authorized command creates the job", update: commandUpdate("/run_dependency_report", 123), wantCalls: 1, wantMessage: "🚀 Job *dependency-report-manual-1791763200* created from *dependency-report*."},
		{name: "unauthorized chat is rejected", update: commandUpdate("/run_dependency_report", 999), wantMessage: "not authorized"},
		{name: "API error is reported", update: commandUpdate("/run_dependency_report", 123), createErr: errors.New("create job: status 403 Forbidden"), wantCalls: 1, wantMessage: "❌ Failed to create the job. Check the hub logs or try again later."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := &fakeJobCreator{err: tt.createErr}
			messenger := &fakeMessenger{}
			handler := NewBotHandler(BotHandlerDeps{TelegramClient: messenger, JobClient: creator, Jobs: []config.JobCommandConfig{job}, Logger: zap.NewNop()})
			handler.Handle(tt.update)
			if len(creator.cronJobs) != tt.wantCalls {
				t.Errorf("create calls = %d, want %d", len(creator.cronJobs), tt.wantCalls)
			}
			if tt.wantCalls > 0 && (creator.namespaces[0] != "internal-apps" || creator.cronJobs[0] != "dependency-report") {
				t.Errorf("created from %v/%v", creator.namespaces, creator.cronJobs)
			}
			messages := messenger.snapshot()
			if len(messages) != 1 || !strings.Contains(messages[0], tt.wantMessage) {
				t.Errorf("messages = %#v, want one containing %q", messages, tt.wantMessage)
			}
		})
	}
}

func TestBotHandlerRunScheduledJob(t *testing.T) {
	job := config.JobCommandConfig{
		Command: "run_dependency_report", Namespace: "internal-apps", CronJob: "dependency-report", AllowedChatIDs: []string{"123", "456"},
	}
	t.Run("success creates the job and stays silent", func(t *testing.T) {
		creator, messenger := &fakeJobCreator{}, &fakeMessenger{}
		NewBotHandler(BotHandlerDeps{TelegramClient: messenger, JobClient: creator, Jobs: []config.JobCommandConfig{job}, Logger: zap.NewNop()}).RunScheduledJob(job)
		if len(creator.cronJobs) != 1 || len(messenger.snapshot()) != 0 {
			t.Errorf("creates = %v, messages = %v", creator.cronJobs, messenger.snapshot())
		}
	})
	t.Run("failure notifies every allowed chat", func(t *testing.T) {
		creator, messenger := &fakeJobCreator{err: errors.New("boom")}, &fakeMessenger{}
		NewBotHandler(BotHandlerDeps{TelegramClient: messenger, JobClient: creator, Jobs: []config.JobCommandConfig{job}, Logger: zap.NewNop()}).RunScheduledJob(job)
		messages := messenger.snapshot()
		if len(messages) != 2 || !strings.Contains(messages[0], "run_dependency_report") {
			t.Errorf("messages = %#v, want one failure per allowed chat naming the command", messages)
		}
	})
}
