package handlers

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"automation-hub/internal/config"
)

type fakeDispatcher struct {
	calls     int
	err       error
	runURL    string
	runURLErr error
}

func (f *fakeDispatcher) DispatchWorkflow(_ context.Context, owner, repo, workflowFile, ref string) error {
	f.calls++
	return f.err
}

func (f *fakeDispatcher) FindLatestRunURL(_ context.Context, owner, repo, workflowFile string, after time.Time) (string, error) {
	return f.runURL, f.runURLErr
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

type fakeMessenger struct {
	mu       sync.Mutex
	messages []string
}

func (f *fakeMessenger) SendMessage(_, message string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
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
			handler := NewBotHandler(dispatcher, nil, messenger, []config.WorkflowCommandConfig{workflow}, nil, zap.NewNop())
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
	handler := NewBotHandler(dispatcher, nil, messenger, []config.WorkflowCommandConfig{workflow}, nil, zap.NewNop())

	handler.Handle(commandUpdate("/run_check_updates", 123))

	if len(messenger.messages) != 1 || !strings.Contains(messenger.messages[0], dispatcher.runURL) {
		t.Fatalf("messages = %#v, want one containing the dispatched run URL %q", messenger.messages, dispatcher.runURL)
	}
	if strings.Contains(messenger.messages[0], "/actions\n") || strings.HasSuffix(messenger.messages[0], "/actions") {
		t.Errorf("message links to the general Actions tab instead of the dispatched run: %q", messenger.messages[0])
	}
}

func TestBotHandlerDispatchWorkflowFallsBackWhenRunLookupFails(t *testing.T) {
	workflow := config.WorkflowCommandConfig{
		Command: "run_check_updates", Owner: "devidence-dev", Repo: "github-runner",
		WorkflowFile: "check-updates.yml", Ref: "main", AllowedChatIDs: []string{"123"},
	}
	dispatcher := &fakeDispatcher{runURLErr: context.DeadlineExceeded}
	messenger := &fakeMessenger{}
	handler := NewBotHandler(dispatcher, nil, messenger, []config.WorkflowCommandConfig{workflow}, nil, zap.NewNop())

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
			handler := NewBotHandler(nil, restarter, messenger, nil, []config.RestartCommandConfig{restart}, zap.NewNop())
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
	handler := NewBotHandler(nil, restarter, messenger, nil, []config.RestartCommandConfig{restart}, zap.NewNop())

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
	handler := NewBotHandler(nil, restarter, messenger, nil, []config.RestartCommandConfig{restart}, zap.NewNop())

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
