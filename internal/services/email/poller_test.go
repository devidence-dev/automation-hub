package email

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"automation-hub/internal/models"
)

// fakeSource offers one email to the handler, and returns err after that.
type fakeSource struct {
	email   models.Email
	err     error
	senders []string
	marked  []bool
}

func (s *fakeSource) Unread(senders []string, handle func(models.Email) bool) error {
	s.senders = senders
	s.marked = append(s.marked, handle(s.email))
	return s.err
}

// fakeDispatcher returns a fixed outcome for every email.
type fakeDispatcher struct {
	senders []string
	result  models.EmailProcessor
	err     error
}

func (d *fakeDispatcher) Senders() []string { return d.senders }

func (d *fakeDispatcher) Dispatch(models.Email) (models.EmailProcessor, error) {
	return d.result, d.err
}

// markingProcessor is an EmailProcessor that implements models.ReadMarker.
type markingProcessor struct{ mark bool }

func (markingProcessor) ShouldProcess(models.Email) bool { return true }
func (markingProcessor) Process(models.Email) error      { return nil }
func (markingProcessor) GetSender() string               { return "" }
func (m markingProcessor) ShouldMarkAsRead() bool        { return m.mark }

// plainProcessor is an EmailProcessor without ShouldMarkAsRead.
type plainProcessor struct{}

func (plainProcessor) ShouldProcess(models.Email) bool { return true }
func (plainProcessor) Process(models.Email) error      { return nil }
func (plainProcessor) GetSender() string               { return "" }

func TestPollerMarkAsReadPolicy(t *testing.T) {
	boom := errors.New("telegram unavailable")
	tests := []struct {
		name       string
		dispatcher *fakeDispatcher
		want       bool
	}{
		{name: "mark_as_read true marks", dispatcher: &fakeDispatcher{result: markingProcessor{mark: true}}, want: true},
		{name: "mark_as_read false does not mark", dispatcher: &fakeDispatcher{result: markingProcessor{mark: false}}, want: false},
		{name: "processor without ShouldMarkAsRead does not mark", dispatcher: &fakeDispatcher{result: plainProcessor{}}, want: false},
		{name: "dispatch error does not mark", dispatcher: &fakeDispatcher{result: markingProcessor{mark: true}, err: boom}, want: false},
		{name: "no match does not mark", dispatcher: &fakeDispatcher{}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poller := NewPoller(&fakeSource{}, tt.dispatcher, time.Second, zap.NewNop())
			if got := poller.handle(models.Email{Subject: "subject"}); got != tt.want {
				t.Errorf("handle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPollerLogsPollError(t *testing.T) {
	const host = "imap.test.com"
	core, logs := observer.New(zap.ErrorLevel)
	source := &fakeSource{err: fmt.Errorf("imap %s: login: authentication failed", host)}
	dispatcher := &fakeDispatcher{senders: []string{"a@example.com"}}
	poller := NewPoller(source, dispatcher, time.Second, zap.New(core))

	poller.poll()

	entries := logs.FilterMessage("Email poll failed").All()
	if len(entries) != 1 {
		t.Fatalf("got %d poll errors, want 1", len(entries))
	}
	if msg := fmt.Sprint(entries[0].ContextMap()["error"]); !strings.Contains(msg, host) {
		t.Errorf("logged error %q does not name host %s", msg, host)
	}
	if len(source.senders) != 1 || source.senders[0] != "a@example.com" {
		t.Errorf("source searched senders %q, want the dispatcher's", source.senders)
	}
}

func TestPollerStopsOnCancel(t *testing.T) {
	poller := NewPoller(&fakeSource{}, &fakeDispatcher{}, time.Millisecond, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		poller.Run(ctx)
		close(done)
	}()
	time.Sleep(5 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after the context was canceled")
	}
}
