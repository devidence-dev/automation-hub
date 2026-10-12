package processor

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

// fakeProcessor is an EmailProcessor that matches or not as configured and counts its Process calls.
type fakeProcessor struct {
	sender string
	match  bool
	err    error
	calls  int
}

func (f *fakeProcessor) ShouldProcess(models.Email) bool { return f.match }

func (f *fakeProcessor) Process(models.Email) error {
	f.calls++
	return f.err
}

func (f *fakeProcessor) GetSender() string { return f.sender }

func TestProcessorManager(t *testing.T) {
	sender := &fakeSender{}
	emailCfg := config.EmailConfig{
		Services: []config.ServiceConfig{
			{Name: "cloudflare", Config: migratedCloudflare},
			{Name: "perplexity", Config: migratedPerplexity},
		},
	}

	mgr, err := NewProcessorManager(emailCfg, sender, zap.NewNop())
	if err != nil {
		t.Fatalf("NewProcessorManager() error = %v", err)
	}
	if got := len(mgr.processors); got != 2 {
		t.Fatalf("Expected 2 processors, got %d", got)
	}

	matching := models.Email{
		From:      "noreply@notify.cloudflare.com",
		Subject:   "Your Verification Code for devidence.dev",
		TextPlain: "Use this code directly: 123456",
	}
	processor, err := mgr.Dispatch(matching)
	if err != nil || processor == nil {
		t.Fatalf("Dispatch(matching) = (%v, %v), want a processor", processor, err)
	}
	if len(sender.messages) != 1 {
		t.Errorf("sent %d messages, want 1", len(sender.messages))
	}

	unrelated := models.Email{From: "unrelated@spam.com", Subject: "Buy drugs online", TextPlain: "Click here"}
	if processor, err := mgr.Dispatch(unrelated); processor != nil || err != nil {
		t.Errorf("Dispatch(unrelated) = (%v, %v), want (nil, nil)", processor, err)
	}
	if len(sender.messages) != 1 {
		t.Errorf("unrelated email sent a message: %q", sender.messages)
	}
}

// TestManagerDispatch_LogsSubjectWhenSendFails checks that a failed send is logged with the email subject.
func TestManagerDispatch_LogsSubjectWhenSendFails(t *testing.T) {
	boom := errors.New("telegram unavailable")
	core, logs := observer.New(zap.ErrorLevel)
	emailCfg := config.EmailConfig{
		Services: []config.ServiceConfig{{Name: "cloudflare", Config: migratedCloudflare}},
	}
	mgr, err := NewProcessorManager(emailCfg, &fakeSender{err: boom}, zap.New(core))
	if err != nil {
		t.Fatalf("NewProcessorManager() error = %v", err)
	}

	const subject = "Your Verification Code for devidence.dev"
	processor, err := mgr.Dispatch(models.Email{
		From:      "noreply@notify.cloudflare.com",
		Subject:   subject,
		TextPlain: "Use this code directly: 123456",
	})
	if !errors.Is(err, boom) || processor == nil {
		t.Fatalf("Dispatch() = (%v, %v), want the processor and the send error", processor, err)
	}

	entries := logs.FilterMessage("Failed to process email").All()
	if len(entries) != 1 {
		t.Fatalf("got %d failure logs, want 1", len(entries))
	}
	if got := entries[0].ContextMap()["subject"]; got != subject {
		t.Errorf("logged subject = %v, want %q", got, subject)
	}
}

func TestNewProcessorManager_InvalidPatternFails(t *testing.T) {
	emailCfg := config.EmailConfig{
		Services: []config.ServiceConfig{
			{Name: "github", Config: config.ServiceProcessorConfig{CodePattern: "[invalid regex ("}},
		},
	}
	if _, err := NewProcessorManager(emailCfg, nil, zap.NewNop()); err == nil {
		t.Error("NewProcessorManager() error = nil, want an error")
	}
}

// dispatchCase is one Dispatch scenario: the processors in config order and what is expected.
type dispatchCase struct {
	name       string
	processors []*fakeProcessor
	wantIndex  int // -1 when no processor is expected
	wantErr    error
	wantCalls  []int
}

func TestManagerDispatch(t *testing.T) {
	boom := errors.New("send failed")
	tests := []dispatchCase{
		{
			name:       "first match handles the email, the second is not called",
			processors: []*fakeProcessor{{match: true}, {match: true}},
			wantIndex:  0,
			wantCalls:  []int{1, 0},
		},
		{
			name:       "unmatched email returns nil and no error",
			processors: []*fakeProcessor{{match: false}, {match: false}},
			wantIndex:  -1,
			wantCalls:  []int{0, 0},
		},
		{
			name:       "failing Process returns its error",
			processors: []*fakeProcessor{{match: true, err: boom}, {match: true}},
			wantIndex:  0,
			wantErr:    boom,
			wantCalls:  []int{1, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &Manager{logger: zap.NewNop()}
			for _, p := range tt.processors {
				mgr.processors = append(mgr.processors, p)
			}

			got, err := mgr.Dispatch(models.Email{Subject: "subject"})
			checkDispatch(t, tt, got, err)
		})
	}
}

// checkDispatch compares the outcome of Dispatch with the expectations of the case.
func checkDispatch(t *testing.T, tt dispatchCase, got models.EmailProcessor, err error) {
	t.Helper()
	if !errors.Is(err, tt.wantErr) {
		t.Errorf("Dispatch() error = %v, want %v", err, tt.wantErr)
	}
	if tt.wantIndex == -1 && got != nil {
		t.Errorf("Dispatch() processor = %v, want nil", got)
	}
	if tt.wantIndex != -1 && got != models.EmailProcessor(tt.processors[tt.wantIndex]) {
		t.Errorf("Dispatch() processor = %v, want the processor at index %d", got, tt.wantIndex)
	}
	for i, p := range tt.processors {
		if p.calls != tt.wantCalls[i] {
			t.Errorf("processor %d called %d times, want %d", i, p.calls, tt.wantCalls[i])
		}
	}
}

func TestManagerSenders(t *testing.T) {
	mgr := &Manager{
		processors: []models.EmailProcessor{
			&fakeProcessor{sender: "a@example.com"},
			&fakeProcessor{sender: ""},
			&fakeProcessor{sender: "b@example.com"},
		},
		logger: zap.NewNop(),
	}
	got := mgr.Senders()
	if len(got) != 2 || got[0] != "a@example.com" || got[1] != "b@example.com" {
		t.Errorf("Senders() = %q, want [a@example.com b@example.com]", got)
	}
}
