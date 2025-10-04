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
	// Connect to server
	imapClient, err := client.DialTLS(fmt.Sprintf("%s:%d", c.config.Host, c.config.Port), nil)
	if err != nil {
		c.logger.Error("Failed to connect to IMAP server", zap.Error(err))
		return
	}
	defer func(imapClient *client.Client) {
		err := imapClient.Logout()
		if err != nil {
			c.logger.Error("Failed to logout from IMAP server", zap.Error(err))
		}
	}(imapClient)

	// Login
	if err := imapClient.Login(c.config.Username, c.config.Password); err != nil {
		c.logger.Error("Failed to login", zap.Error(err))
		return
	}

	// Select INBOX
	_, err = imapClient.Select("INBOX", false)
	if err != nil {
		c.logger.Error("Failed to select INBOX", zap.Error(err))
		return
	}

	// Search for unread messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	ids, err := imapClient.Search(criteria)
	if err != nil {
		c.logger.Error("Failed to search emails", zap.Error(err))
		return
	}

	if len(ids) == 0 {
		return
	}

	c.logger.Info("Found unread emails", zap.Int("count", len(ids)))

	// Fetch messages
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	messages := make(chan *imap.Message, len(ids))
	done := make(chan error, 1)

	go func() {
		done <- imapClient.Fetch(seqset, []imap.FetchItem{
			imap.FetchEnvelope,
			imap.FetchBodyStructure,
			"BODY[TEXT]", // Fetch the text content specifically
			imap.FetchFlags,
			imap.FetchUid,
		}, messages)
	}()

	for msg := range messages {
		email := c.parseMessage(msg)

		// Process with each processor
		processed := false
		for _, processor := range processors {
			if processor.ShouldProcess(email) {
				c.logger.Info("Processing email",
					zap.String("subject", email.Subject),
					zap.String("from", email.From))
				if err := processor.Process(email); err != nil {
					c.logger.Error("Failed to process email",
						zap.String("subject", email.Subject),
						zap.Error(err))
				} else {
					c.logger.Info("Email processed successfully",
						zap.String("subject", email.Subject))
					// Mark as read
					seqSet := new(imap.SeqSet)
					seqSet.AddNum(msg.SeqNum)
					flags := []interface{}{imap.SeenFlag}
					if err := imapClient.Store(seqSet, "+FLAGS", flags, nil); err != nil {
						c.logger.Error("Failed to mark email as read", zap.Error(err))
					}
				}
				processed = true
				break
			}
		}

		if !processed {
			c.logger.Info("Email ignored (no matching processor)",
				zap.String("subject", email.Subject),
				zap.String("from", email.From))
		}
	}

	if err := <-done; err != nil {
		c.logger.Error("Failed to fetch messages", zap.Error(err))
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
