package handlers

import (
	"context"
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"automation-hub/internal/config"
)

type fakeDispatcher struct {
	calls int
	err   error
}

func (f *fakeDispatcher) DispatchWorkflow(_ context.Context, _, _, _, _ string) error {
	f.calls++
	return f.err
}

type fakeMessenger struct {
	messages []string
}

func (f *fakeMessenger) SendMessage(_ string, message string) error {
	f.messages = append(f.messages, message)
	return nil
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
		{name: "authorized command dispatches workflow", update: commandUpdate("/run_check_updates", 123), wantCalls: 1, wantMessage: "disparado correctamente"},
		{name: "unauthorized chat is rejected", update: commandUpdate("/run_check_updates", 999), wantMessage: "No estás autorizado"},
		{name: "unknown command does not dispatch", update: commandUpdate("/run_deploy_prod", 123), wantMessage: "Comando no reconocido"},
		{name: "github failure is reported", update: commandUpdate("/run_check_updates", 123), dispatchError: context.DeadlineExceeded, wantCalls: 1, wantMessage: "No se pudo disparar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher := &fakeDispatcher{err: tt.dispatchError}
			messenger := &fakeMessenger{}
			handler := NewBotHandler(dispatcher, messenger, []config.WorkflowCommandConfig{workflow}, zap.NewNop())
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

func commandUpdate(command string, chatID int64) tgbotapi.Update {
	return tgbotapi.Update{Message: &tgbotapi.Message{
		Text:     command,
		Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(command)}},
		Chat:     &tgbotapi.Chat{ID: chatID},
	}}
}
