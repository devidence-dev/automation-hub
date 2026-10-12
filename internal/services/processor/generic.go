package processor

import (
	"fmt"
	"io"
	"mime/quotedprintable"
	"strings"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
)

// GenericEmailProcessor forwards the code of the emails one service matches. It composes a matcher,
// body preparation, a CodeExtractor and a TemplateNotifier.
type GenericEmailProcessor struct {
	name         string
	config       config.ServiceProcessorConfig
	extractor    CodeExtractor
	notifier     TemplateNotifier
	logger       *zap.Logger
	notFoundText string
}

// The poller finds mark-as-read with a runtime type assertion to models.ReadMarker.
var _ models.ReadMarker = (*GenericEmailProcessor)(nil)

// NewGenericEmailProcessor builds the processor of one service. logPreviewLength caps the text shown in
// not-found warnings. It fails when the service's regexes do not compile.
func NewGenericEmailProcessor(name string, serviceConfig config.ServiceProcessorConfig, logPreviewLength int, sender messageSender, logger *zap.Logger) (*GenericEmailProcessor, error) {
	logger = logger.With(zap.String("service", name))
	extractor, err := newCodeExtractor(serviceConfig, logPreviewLength, logger)
	if err != nil {
		return nil, fmt.Errorf("service %q: %w", name, err)
	}
	return &GenericEmailProcessor{
		name:      name,
		config:    serviceConfig,
		extractor: extractor,
		notifier: TemplateNotifier{
			sender:   sender,
			chatID:   serviceConfig.TelegramChatID,
			template: serviceConfig.TelegramMessage,
		},
		logger:       logger,
		notFoundText: serviceConfig.EffectiveNotFoundText(),
	}, nil
}

// ShouldProcess is the matcher: the sender must contain email_from, and the subject one of email_subject.
func (p *GenericEmailProcessor) ShouldProcess(email models.Email) bool {
	// Check the sender
	if !strings.Contains(email.From, p.config.EmailFrom) {
		return false
	}

	// Check at least one of the subjects
	for _, subject := range p.config.EmailSubject {
		if strings.Contains(email.Subject, subject) {
			return true
		}
	}

	return false
}

func (p *GenericEmailProcessor) Process(email models.Email) error {
	p.logger.Debug("Processing email",
		zap.String("from", email.From),
		zap.String("subject", email.Subject))
	return p.notifier.Notify(p.extractCode(email.TextPlain))
}

func (p *GenericEmailProcessor) GetName() string {
	return p.name
}

func (p *GenericEmailProcessor) GetSender() string {
	return p.config.EmailFrom
}

func (p *GenericEmailProcessor) ShouldMarkAsRead() bool {
	return p.config.ShouldMarkAsRead()
}

// extractCode prepares the body, extracts the code and substitutes the not-found text when none is found.
func (p *GenericEmailProcessor) extractCode(text string) string {
	body := p.prepareBody(text)
	p.logger.Debug("Email content prepared", zap.String("decoded_text", body))
	code, found := p.extractor.Extract(body)
	if !found {
		return p.notFoundText
	}
	return code
}

// prepareBody decodes quoted-printable content, then removes the MIME headers unless the service keeps them.
func (p *GenericEmailProcessor) prepareBody(text string) string {
	body := p.decodeQuotedPrintable(text)
	if p.config.EffectiveStripMIMEHeaders() {
		body = p.stripMIMEHeaders(body)
	}
	return body
}

func (p *GenericEmailProcessor) stripMIMEHeaders(text string) string {
	// Find the first blank line (end of headers)
	headerEnd := strings.Index(text, "\n\n")
	if headerEnd == -1 {
		headerEnd = strings.Index(text, "\r\n\r\n")
		if headerEnd == -1 {
			return text // No headers found, return as is
		}
		headerEnd += 4 // Include the \r\n\r\n
	} else {
		headerEnd += 2 // Include the \n\n
	}
	return text[headerEnd:]
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (p *GenericEmailProcessor) decodeQuotedPrintable(text string) string {
	// If the text contains quoted-printable characters, try to decode it
	if strings.Contains(text, "=") {
		reader := quotedprintable.NewReader(strings.NewReader(text))
		if decoded, err := io.ReadAll(reader); err == nil {
			return string(decoded)
		} else {
			p.logger.Warn("Failed to decode quoted-printable", zap.Error(err))
		}
	}
	return text
}
