package processor

import "testing"

// fakeSender records every message it is asked to send.
type fakeSender struct {
	chatIDs  []string
	messages []string
	err      error
}

func (f *fakeSender) SendMessage(chatID, message string) error {
	f.chatIDs = append(f.chatIDs, chatID)
	f.messages = append(f.messages, message)
	return f.err
}

func TestTemplateNotifier(t *testing.T) {
	tests := []struct {
		name     string
		template string
		args     []any
		want     string
	}{
		{name: "one argument", template: "Code: %s", args: []any{"123456"}, want: "Code: 123456"},
		{name: "two arguments", template: "Download finished: %s at %s", args: []any{"Dune", "/movies"}, want: "Download finished: Dune at /movies"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeSender{}
			notifier := TemplateNotifier{sender: sender, chatID: "42", template: tt.template}
			if err := notifier.Notify(tt.args...); err != nil {
				t.Fatalf("Notify() error = %v", err)
			}
			if len(sender.messages) != 1 || sender.messages[0] != tt.want {
				t.Errorf("sent %q, want %q", sender.messages, tt.want)
			}
			if len(sender.chatIDs) != 1 || sender.chatIDs[0] != "42" {
				t.Errorf("chat IDs = %q, want [42]", sender.chatIDs)
			}
		})
	}
}
