package email

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/emersion/go-imap"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

// fakeSession is an imapSession with scripted search results, fetched messages and failures.
type fakeSession struct {
	loginErr  error
	selectErr error
	searchErr map[string]error    // by sender; "" for the search of all unread mail
	results   map[string][]uint32 // by sender; "" for the search of all unread mail
	fetched   []*imap.Message
	fetchErr  error
	stores    int
	logouts   int
}

func (f *fakeSession) Login(string, string) error { return f.loginErr }

func (f *fakeSession) Select(string, bool) (*imap.MailboxStatus, error) { return nil, f.selectErr }

func (f *fakeSession) Search(criteria *imap.SearchCriteria) ([]uint32, error) {
	sender := criteria.Header.Get("From")
	if err := f.searchErr[sender]; err != nil {
		return nil, err
	}
	return f.results[sender], nil
}

// Fetch sends the fetched messages whose sequence number was requested, as a server does.
func (f *fakeSession) Fetch(seqset *imap.SeqSet, _ []imap.FetchItem, ch chan *imap.Message) error {
	defer close(ch)
	for _, msg := range f.fetched {
		if inSeqSet(seqset, msg.SeqNum) {
			ch <- msg
		}
	}
	return f.fetchErr
}

// inSeqSet reports whether num is in set.
func inSeqSet(set *imap.SeqSet, num uint32) bool {
	for _, seq := range set.Set {
		if seq.Start <= num && num <= seq.Stop {
			return true
		}
	}
	return false
}

func (f *fakeSession) Store(*imap.SeqSet, imap.StoreItem, interface{}, chan *imap.Message) error {
	f.stores++
	return nil
}

func (f *fakeSession) Logout() error {
	f.logouts++
	return nil
}

// newFakeSource returns an IMAPSource whose dial always gives session.
func newFakeSource(session *fakeSession, logger *zap.Logger) *IMAPSource {
	return &IMAPSource{
		config: config.EmailConfig{Host: "imap.test.com", Port: 993, Username: "user", Password: "pass"},
		dial: func(string) (imapSession, error) {
			return session, nil
		},
		logger: logger,
	}
}

// messageFrom returns a fetched message whose envelope is from the given address.
func messageFrom(seqNum uint32, mailbox, host string) *imap.Message {
	return &imap.Message{
		SeqNum: seqNum,
		Envelope: &imap.Envelope{
			MessageId: fmt.Sprintf("<%d@test>", seqNum),
			Subject:   "Your login code",
			From:      []*imap.Address{{MailboxName: mailbox, HostName: host}},
		},
	}
}

func TestNewIMAPSource(t *testing.T) {
	source := NewIMAPSource(config.EmailConfig{Host: "imap.test.com", Port: 993}, zap.NewNop())
	if source == nil || source.dial == nil {
		t.Fatal("Expected NewIMAPSource to return a source with a dialer")
	}
}

func TestUnreadOneSenderFails(t *testing.T) {
	const cloudflare = "noreply@notify.cloudflare.com"
	const openai = "noreply@tm.openai.com"
	session := &fakeSession{
		searchErr: map[string]error{openai: errors.New("search failed")},
		results:   map[string][]uint32{cloudflare: {1}},
		fetched:   []*imap.Message{messageFrom(1, "noreply", "notify.cloudflare.com")},
	}
	logger, logs := observer.New(zap.ErrorLevel)
	source := newFakeSource(session, zap.New(logger))

	var offered []models.Email
	err := source.Unread([]string{openai, cloudflare}, func(email models.Email) bool {
		offered = append(offered, email)
		return true
	})
	if err != nil {
		t.Fatalf("Unread() error = %v, want nil when one sender succeeds", err)
	}
	if len(offered) != 1 || offered[0].From != cloudflare {
		t.Errorf("offered %+v, want the Cloudflare email only", offered)
	}
	if session.stores != 1 {
		t.Errorf("stores = %d, want 1 (handle returned true)", session.stores)
	}
	entries := logs.FilterMessage("Failed to search emails for sender").All()
	if len(entries) != 1 || entries[0].ContextMap()["sender"] != openai {
		t.Errorf("want one search error logged with sender %s, got %+v", openai, entries)
	}
}

// TestUnreadTwoServicesOneUnrelated checks that two senders' searches offer each email once, when one ID
// is returned by both searches. The unrelated email (ID 4) is not returned by either search.
func TestUnreadTwoServicesOneUnrelated(t *testing.T) {
	const cloudflare = "noreply@notify.cloudflare.com"
	const openai = "noreply@tm.openai.com"
	session := &fakeSession{
		results: map[string][]uint32{cloudflare: {1, 2}, openai: {2, 3}},
		fetched: []*imap.Message{
			messageFrom(1, "noreply", "notify.cloudflare.com"),
			messageFrom(2, "noreply", "notify.cloudflare.com"),
			messageFrom(3, "noreply", "tm.openai.com"),
			messageFrom(4, "news", "example.org"),
		},
	}
	source := newFakeSource(session, zap.NewNop())

	offered := make(map[string]int)
	err := source.Unread([]string{cloudflare, openai}, func(email models.Email) bool {
		offered[email.ID]++
		return true
	})
	if err != nil {
		t.Fatalf("Unread() error = %v", err)
	}
	want := []string{"<1@test>", "<2@test>", "<3@test>"}
	if len(offered) != len(want) {
		t.Errorf("offered %v, want exactly %v", offered, want)
	}
	for _, id := range want {
		if offered[id] != 1 {
			t.Errorf("email %s offered %d times, want 1", id, offered[id])
		}
	}
	if session.stores != len(want) {
		t.Errorf("stores = %d, want %d (handle returned true for each offered email)", session.stores, len(want))
	}
}

func TestUnreadAllSendersFail(t *testing.T) {
	session := &fakeSession{
		searchErr: map[string]error{
			"a@example.com": errors.New("search failed"),
			"b@example.com": errors.New("search failed"),
		},
	}
	source := newFakeSource(session, zap.NewNop())

	called := false
	err := source.Unread([]string{"a@example.com", "b@example.com"}, func(models.Email) bool {
		called = true
		return false
	})
	if err == nil || !strings.Contains(err.Error(), "imap.test.com") {
		t.Errorf("Unread() error = %v, want one naming the host", err)
	}
	if called {
		t.Error("handle was called although every search failed")
	}
}

func TestUnreadLoginFails(t *testing.T) {
	session := &fakeSession{loginErr: errors.New("authentication failed")}
	source := newFakeSource(session, zap.NewNop())

	err := source.Unread([]string{"a@example.com"}, func(models.Email) bool { return true })
	if err == nil || !strings.Contains(err.Error(), "imap.test.com") {
		t.Errorf("Unread() error = %v, want one naming the host", err)
	}
	if session.logouts != 1 {
		t.Errorf("logouts = %d, want 1 after the failed login", session.logouts)
	}
}

func TestUnreadWithoutSendersSearchesAll(t *testing.T) {
	session := &fakeSession{
		results: map[string][]uint32{"": {7}},
		fetched: []*imap.Message{messageFrom(7, "noreply", "example.com")},
	}
	source := newFakeSource(session, zap.NewNop())

	var offered int
	if err := source.Unread(nil, func(models.Email) bool { offered++; return false }); err != nil {
		t.Fatalf("Unread() error = %v", err)
	}
	if offered != 1 || session.stores != 0 {
		t.Errorf("offered = %d, stores = %d, want 1 and 0 (handle returned false)", offered, session.stores)
	}
}

// TestParseMessageNilEnvelope checks that a message without an envelope parses with empty fields and no panic.
func TestParseMessageNilEnvelope(t *testing.T) {
	source := newFakeSource(&fakeSession{}, zap.NewNop())
	section, err := imap.ParseBodySectionName("BODY[TEXT]")
	if err != nil {
		t.Fatalf("ParseBodySectionName() error = %v", err)
	}
	msg := &imap.Message{Body: map[*imap.BodySectionName]imap.Literal{section: bytes.NewBufferString("123456")}}

	email := source.parseMessage(msg)
	if email.ID != "" || email.Subject != "" || email.From != "" {
		t.Errorf("email = %+v, want empty ID, subject and sender", email)
	}
	if email.TextPlain != "123456" {
		t.Errorf("TextPlain = %q, want 123456", email.TextPlain)
	}
}

func TestExtractTextPlain(t *testing.T) {
	source := newFakeSource(&fakeSession{}, zap.NewNop())

	if got := source.extractTextPlain(nil); got != "" {
		t.Errorf("Expected empty string for nil body, got %q", got)
	}

	buf := bytes.NewBufferString("Hello World Body")
	if got := source.extractTextPlain(buf); got != "Hello World Body" {
		t.Errorf("extractTextPlain() = %q, expected %q", got, "Hello World Body")
	}
}

func TestParseMessage(t *testing.T) {
	source := newFakeSource(&fakeSession{}, zap.NewNop())

	msg := &imap.Message{
		Envelope: &imap.Envelope{
			MessageId: "msg-123",
			Subject:   "Test Email Subject",
			From: []*imap.Address{
				{
					PersonalName: "Sender Name",
					MailboxName:  "support",
					HostName:     "service.com",
				},
			},
		},
	}

	email := source.parseMessage(msg)
	if email.ID != "msg-123" {
		t.Errorf("Expected ID msg-123, got %s", email.ID)
	}
	if email.Subject != "Test Email Subject" {
		t.Errorf("Expected Subject Test Email Subject, got %s", email.Subject)
	}
	if email.From != "support@service.com" {
		t.Errorf("Expected From support@service.com, got %s", email.From)
	}
}

func TestMarkAsReadNilSession(t *testing.T) {
	source := newFakeSource(&fakeSession{}, zap.NewNop())

	source.markAsRead(nil, 1)
}
