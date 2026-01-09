package email

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

type IMAPClient struct {
	config config.EmailConfig
	logger *zap.Logger
}

func NewIMAPClient(config config.EmailConfig, logger *zap.Logger) *IMAPClient {
	return &IMAPClient{
		config: config,
		logger: logger,
	}
}

func (c *IMAPClient) StartMonitoring(ctx context.Context, processors ...models.EmailProcessor) {
	// Use the polling interval from the configuration, default 60 seconds if not configured
	pollingInterval := time.Duration(c.config.PollingInterval) * time.Second
	if c.config.PollingInterval == 0 {
		pollingInterval = 60 * time.Second
	}

	c.logger.Info("Starting email monitoring",
		zap.Duration("polling_interval", pollingInterval))

	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.checkEmails(processors...)
		}
	}
}

func (c *IMAPClient) checkEmails(processors ...models.EmailProcessor) {
	imapClient, err := c.connectAndLogin()
	if err != nil {
		return
	}
	defer c.logout(imapClient)

	var senders []string
	for _, p := range processors {
		if s := p.GetSender(); s != "" {
			senders = append(senders, s)
		}
	}

	ids, err := c.searchUnreadEmails(imapClient, senders)
	if err != nil {
		return
	}

	if len(ids) == 0 {
		return
	}

	c.fetchAndProcessMessages(imapClient, ids, processors...)
}

func (c *IMAPClient) connectAndLogin() (*client.Client, error) {
	imapClient, err := client.DialTLS(fmt.Sprintf("%s:%d", c.config.Host, c.config.Port), nil)
	if err != nil {
		c.logger.Error("Failed to connect to IMAP server", zap.Error(err))
		return nil, err
	}

	if err := imapClient.Login(c.config.Username, c.config.Password); err != nil {
		c.logger.Error("Failed to login", zap.Error(err))
		if logoutErr := imapClient.Logout(); logoutErr != nil {
			c.logger.Error("Failed to logout after login failure", zap.Error(logoutErr))
		}
		return nil, err
	}

	_, err = imapClient.Select("INBOX", false)
	if err != nil {
		c.logger.Error("Failed to select INBOX", zap.Error(err))
		if logoutErr := imapClient.Logout(); logoutErr != nil {
			c.logger.Error("Failed to logout after select failure", zap.Error(logoutErr))
		}
		return nil, err
	}

	return imapClient, nil
}

func (c *IMAPClient) logout(imapClient *client.Client) {
	if err := imapClient.Logout(); err != nil {
		c.logger.Error("Failed to logout from IMAP server", zap.Error(err))
	}
}

func (c *IMAPClient) searchUnreadEmails(imapClient *client.Client, senders []string) ([]uint32, error) {
	if len(senders) == 0 {
		// Fallback to searching all unread emails if no senders specified
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{imap.SeenFlag}

		ids, err := imapClient.Search(criteria)
		if err != nil {
			c.logger.Error("Failed to search emails", zap.Error(err))
			return nil, err
		}
		if len(ids) > 0 {
			c.logger.Info("Found unread emails (all sources)", zap.Int("count", len(ids)))
		}
		return ids, nil
	}

	uniqueIDs := make(map[uint32]struct{})
	for _, sender := range senders {
		criteria := imap.NewSearchCriteria()
		criteria.WithoutFlags = []string{imap.SeenFlag}
		criteria.Header.Add("From", sender)

		ids, err := imapClient.Search(criteria)
		if err != nil {
			c.logger.Error("Failed to search emails for sender", zap.String("sender", sender), zap.Error(err))
			continue
		}
		for _, id := range ids {
			uniqueIDs[id] = struct{}{}
		}
	}

	var allIDs []uint32
	for id := range uniqueIDs {
		allIDs = append(allIDs, id)
	}

	if len(allIDs) > 0 {
		c.logger.Info("Found unread emails from allowed senders", zap.Int("count", len(allIDs)))
	}

	return allIDs, nil
}

func (c *IMAPClient) fetchAndProcessMessages(imapClient *client.Client, ids []uint32, processors ...models.EmailProcessor) {
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	messages := make(chan *imap.Message, len(ids))
	done := make(chan error, 1)

	go func() {
		done <- imapClient.Fetch(seqset, []imap.FetchItem{
			imap.FetchEnvelope,
			imap.FetchBodyStructure,
			"BODY.PEEK[TEXT]",
			imap.FetchFlags,
			imap.FetchUid,
		}, messages)
	}()

	for msg := range messages {
		c.processMessage(imapClient, msg, processors...)
	}

	if err := <-done; err != nil {
		c.logger.Error("Failed to fetch messages", zap.Error(err))
	}
}

func (c *IMAPClient) processMessage(imapClient *client.Client, msg *imap.Message, processors ...models.EmailProcessor) {
	email := c.parseMessage(msg)

	for _, processor := range processors {
		if processor.ShouldProcess(email) {
			c.logger.Info("Processing email",
				zap.String("subject", email.Subject),
				zap.String("from", email.From))

			// Try to process the email
			if err := processor.Process(email); err != nil {
				c.logger.Error("Failed to process email",
					zap.String("subject", email.Subject),
					zap.String("from", email.From),
					zap.Error(err))
				return
			}

			c.logger.Info("Email processed successfully",
				zap.String("subject", email.Subject),
				zap.String("from", email.From))

			// Handle post-processing (marking as read only for Perplexity/Cloudflare)
			c.handlePostProcessing(imapClient, processor, msg, email)
			return
		}
	}

	// No processor matched this email, leave it unread
	c.logger.Info("Email ignored (no matching processor)",
		zap.String("subject", email.Subject),
		zap.String("from", email.From))
}

func (c *IMAPClient) handlePostProcessing(imapClient *client.Client, processor models.EmailProcessor, msg *imap.Message, email models.Email) {
	named, ok := processor.(interface{ GetName() string })
	if !ok {
		c.logger.Debug("Processor has no GetName, not marking as read",
			zap.String("subject", email.Subject))
		return
	}

	name := strings.ToLower(named.GetName())
	// Only mark as read for Perplexity and Cloudflare processors
	if name == "perplexity" || name == "cloudflare" {
		c.logger.Info("Marking email as read (whitelisted processor)",
			zap.String("processor", name),
			zap.String("from", email.From),
			zap.String("subject", email.Subject))
		c.markAsRead(imapClient, msg.SeqNum)
	} else {
		c.logger.Info("Email processed but NOT marked as read (processor not whitelisted)",
			zap.String("processor", name),
			zap.String("from", email.From),
			zap.String("subject", email.Subject))
	}
}

func (c *IMAPClient) markAsRead(imapClient *client.Client, seqNum uint32) {
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNum)
	flags := []interface{}{imap.SeenFlag}
	if err := imapClient.Store(seqSet, "+FLAGS", flags, nil); err != nil {
		c.logger.Error("Failed to mark email as read", zap.Error(err))
	}
}

func (c *IMAPClient) markAsUnread(imapClient *client.Client, seqNum uint32) {
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNum)
	flags := []interface{}{imap.SeenFlag}
	if err := imapClient.Store(seqSet, "-FLAGS", flags, nil); err != nil {
		c.logger.Error("Failed to mark email as unread", zap.Error(err))
	}
}

func (c *IMAPClient) parseMessage(msg *imap.Message) models.Email {
	email := models.Email{
		ID: msg.Envelope.MessageId,
	}

	if msg.Envelope != nil {
		email.Subject = msg.Envelope.Subject
		if len(msg.Envelope.From) > 0 {
			email.From = msg.Envelope.From[0].Address()
		}
	}

	// Parse body - look for BODY[TEXT] section
	for sectionName, body := range msg.Body {
		sectionStr := string(sectionName.FetchItem())
		c.logger.Info("Found email section", zap.String("section", sectionStr))
		if strings.Contains(sectionStr, "TEXT") {
			email.TextPlain = c.extractTextPlain(body)
			break
		}
	}

	return email
}

func (c *IMAPClient) extractTextPlain(body imap.Literal) string {
	if body == nil {
		return ""
	}

	// Read the body content
	buf, err := io.ReadAll(body)
	if err != nil {
		c.logger.Error("Failed to read email body", zap.Error(err))
		return ""
	}

	return string(buf)
}
