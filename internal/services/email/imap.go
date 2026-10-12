package email

import (
	"fmt"
	"io"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

// imapSession is the part of a go-imap connection the source uses. *client.Client satisfies it.
type imapSession interface {
	Login(username, password string) error
	Select(name string, readOnly bool) (*imap.MailboxStatus, error)
	Search(criteria *imap.SearchCriteria) ([]uint32, error)
	Fetch(seqset *imap.SeqSet, items []imap.FetchItem, ch chan *imap.Message) error
	Store(seqset *imap.SeqSet, item imap.StoreItem, value interface{}, ch chan *imap.Message) error
	Logout() error
}

// IMAPSource is the go-imap transport: it connects, searches unread mail per sender, fetches and flags it.
type IMAPSource struct {
	config config.EmailConfig
	dial   func(addr string) (imapSession, error)
	logger *zap.Logger
}

func NewIMAPSource(config config.EmailConfig, logger *zap.Logger) *IMAPSource {
	return &IMAPSource{
		config: config,
		dial:   dialTLS,
		logger: logger,
	}
}

func dialTLS(addr string) (imapSession, error) {
	session, err := client.DialTLS(addr, nil)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Unread connects, logs in, selects INBOX and offers each unread email from senders to handle. Errors
// wrap the IMAP host; the poller logs them.
func (s *IMAPSource) Unread(senders []string, handle func(models.Email) bool) error {
	session, err := s.connect()
	if err != nil {
		return err
	}
	defer s.logout(session)

	ids, err := s.search(session, senders)
	if err != nil || len(ids) == 0 {
		return err
	}
	return s.fetch(session, ids, handle)
}

func (s *IMAPSource) connect() (imapSession, error) {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	session, err := s.dial(addr)
	if err != nil {
		return nil, fmt.Errorf("imap %s: connect: %w", s.config.Host, err)
	}
	if err := session.Login(s.config.Username, s.config.Password); err != nil {
		s.logout(session)
		return nil, fmt.Errorf("imap %s: login: %w", s.config.Host, err)
	}
	if _, err := session.Select("INBOX", false); err != nil {
		s.logout(session)
		return nil, fmt.Errorf("imap %s: select INBOX: %w", s.config.Host, err)
	}
	return session, nil
}

func (s *IMAPSource) logout(session imapSession) {
	if err := session.Logout(); err != nil {
		s.logger.Error("Failed to logout from IMAP server", zap.Error(err))
	}
}

// search returns the IDs of the unread emails from any sender, or of all unread emails when senders is
// empty. A failed search for one sender is logged and skipped; the search fails only when every one fails.
func (s *IMAPSource) search(session imapSession, senders []string) ([]uint32, error) {
	if len(senders) == 0 {
		ids, err := session.Search(unreadCriteria(""))
		if err != nil {
			return nil, fmt.Errorf("imap %s: search unread: %w", s.config.Host, err)
		}
		if len(ids) > 0 {
			s.logger.Info("Found unread emails (all sources)", zap.Int("count", len(ids)))
		}
		return ids, nil
	}

	uniqueIDs := make(map[uint32]struct{})
	var lastErr error
	failed := 0
	for _, sender := range senders {
		ids, err := session.Search(unreadCriteria(sender))
		if err != nil {
			s.logger.Error("Failed to search emails for sender", zap.String("sender", sender), zap.Error(err))
			lastErr = err
			failed++
			continue
		}
		for _, id := range ids {
			uniqueIDs[id] = struct{}{}
		}
	}
	if failed == len(senders) {
		return nil, fmt.Errorf("imap %s: every sender search failed: %w", s.config.Host, lastErr)
	}

	var allIDs []uint32
	for id := range uniqueIDs {
		allIDs = append(allIDs, id)
	}
	if len(allIDs) > 0 {
		s.logger.Info("Found unread emails from allowed senders", zap.Int("count", len(allIDs)))
	}
	return allIDs, nil
}

// unreadCriteria matches unread mail, and mail whose From contains sender when sender is not empty.
func unreadCriteria(sender string) *imap.SearchCriteria {
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	if sender != "" {
		criteria.Header.Add("From", sender)
	}
	return criteria
}

func (s *IMAPSource) fetch(session imapSession, ids []uint32, handle func(models.Email) bool) error {
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	messages := make(chan *imap.Message, len(ids))
	done := make(chan error, 1)

	go func() {
		done <- session.Fetch(seqset, []imap.FetchItem{
			imap.FetchEnvelope,
			imap.FetchBodyStructure,
			"BODY.PEEK[TEXT]",
			imap.FetchFlags,
			imap.FetchUid,
		}, messages)
	}()

	for msg := range messages {
		if handle(s.parseMessage(msg)) {
			s.markAsRead(session, msg.SeqNum)
		}
	}

	if err := <-done; err != nil {
		return fmt.Errorf("imap %s: fetch: %w", s.config.Host, err)
	}
	return nil
}

func (s *IMAPSource) markAsRead(session imapSession, seqNum uint32) {
	if session == nil {
		return
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNum)
	flags := []interface{}{imap.SeenFlag}
	if err := session.Store(seqSet, "+FLAGS", flags, nil); err != nil {
		s.logger.Error("Failed to mark email as read", zap.Error(err))
	}
}

// parseMessage builds the email of a fetched message. A message without an envelope keeps empty fields.
func (s *IMAPSource) parseMessage(msg *imap.Message) models.Email {
	var email models.Email

	if msg.Envelope != nil {
		email.ID = msg.Envelope.MessageId
		email.Subject = msg.Envelope.Subject
		if len(msg.Envelope.From) > 0 {
			email.From = msg.Envelope.From[0].Address()
		}
	}

	// Parse body - look for BODY[TEXT] section
	for sectionName, body := range msg.Body {
		sectionStr := string(sectionName.FetchItem())
		s.logger.Info("Found email section", zap.String("section", sectionStr))
		if strings.Contains(sectionStr, "TEXT") {
			email.TextPlain = s.extractTextPlain(body)
			break
		}
	}

	return email
}

func (s *IMAPSource) extractTextPlain(body imap.Literal) string {
	if body == nil {
		return ""
	}

	// Read the body content
	buf, err := io.ReadAll(body)
	if err != nil {
		s.logger.Error("Failed to read email body", zap.Error(err))
		return ""
	}

	return string(buf)
}
