package processor

import (
	"fmt"
	"io"
	"mime/quotedprintable"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
	"automation-hub/internal/services/telegram"
)

const NotFoundCode = "Not found"

type GenericEmailProcessor struct {
	name            string
	config          config.ServiceProcessorConfig
	telegram        *telegram.Client
	logger          *zap.Logger
	codePattern     *regexp.Regexp
	defaultPatterns map[string]*regexp.Regexp
}

func NewGenericEmailProcessor(name string, serviceConfig config.ServiceProcessorConfig, telegram *telegram.Client, logger *zap.Logger) *GenericEmailProcessor {
	processor := &GenericEmailProcessor{
		name:     name,
		config:   serviceConfig,
		telegram: telegram,
		logger:   logger,
		defaultPatterns: map[string]*regexp.Regexp{
			"cloudflare": regexp.MustCompile(`\b\d{6}\b`),
			// Perplexity pattern is now more flexible - handles both numeric and alphanumeric formats
			"perplexity": regexp.MustCompile(`(?:\d{5,6}|[a-zA-Z0-9]+-[a-zA-Z0-9]+)`),
			"default":    regexp.MustCompile(`\b[a-zA-Z0-9]{4,8}\b`), // generic pattern
		},
	}

	// If there is a custom pattern, use it
	if serviceConfig.CodePattern != "" {
		if pattern, err := regexp.Compile(serviceConfig.CodePattern); err == nil {
			processor.codePattern = pattern
		} else {
			logger.Warn("Invalid custom code pattern, using default",
				zap.String("service", name),
				zap.String("pattern", serviceConfig.CodePattern),
				zap.Error(err))
		}
	}

	// If there is no custom pattern, use the default pattern of the service
	if processor.codePattern == nil {
		if pattern, exists := processor.defaultPatterns[strings.ToLower(name)]; exists {
			processor.codePattern = pattern
		} else {
			processor.codePattern = processor.defaultPatterns["default"]
		}
	}

	return processor
}

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
	// Decode quoted-printable content if necessary
	decodedText := p.decodeQuotedPrintable(email.TextPlain)

	// Log the decoded content for debugging
	p.logger.Debug("Processing email content",
		zap.String("service", p.name),
		zap.String("from", email.From),
		zap.String("subject", email.Subject),
		zap.String("decoded_text", decodedText))

	// Extract the code using the configured pattern
	code := p.extractCode(decodedText)

	// Format the message
	message := fmt.Sprintf(p.config.TelegramMessage, code)

	// Send message to Telegram
	return p.telegram.SendMessage(p.config.TelegramChatID, message)
}

func (p *GenericEmailProcessor) GetName() string {
	return p.name
}

func (p *GenericEmailProcessor) GetSender() string {
	return p.config.EmailFrom
}

func (p *GenericEmailProcessor) extractCode(text string) string {
	var body string
	if strings.ToLower(p.name) == "cloudflare" {
		body = text
	} else {
		body = p.stripMIMEHeaders(text)
	}

	// For Perplexity, find the code after "directamente:"
	if strings.ToLower(p.name) == "perplexity" {
		return p.extractPerplexityCode(body)
	}

	matches := p.codePattern.FindStringSubmatch(body)
	if len(matches) > 0 {
		p.logger.Info("Code extracted successfully",
			zap.String("service", p.name),
			zap.String("code", matches[0]))
		return matches[0]
	}
	p.logger.Warn("Code not found in email",
		zap.String("service", p.name),
		zap.String("pattern", p.codePattern.String()),
		zap.String("text_preview", truncateString(body, 200)))
	return NotFoundCode
}

func (p *GenericEmailProcessor) extractPerplexityCode(text string) string {
	// Find the position after "directamente:" (Spanish) or "directly:" (English)
	markers := []string{"directly:", "directamente:"}
	var searchText string
	var markerFound bool

	lowerText := strings.ToLower(text)
	for _, marker := range markers {
		if idx := strings.Index(lowerText, marker); idx != -1 {
			searchText = text[idx+len(marker):]
			markerFound = true
			p.logger.Debug("Perplexity marker found", zap.String("marker", marker))
			break
		}
	}

	if !markerFound {
		p.logger.Warn("Neither 'directly:' nor 'directamente:' marker found in Perplexity email")
		return NotFoundCode
	}

	// Try multiple patterns in order of preference
	// Pattern 1: Numeric only (5-6 digits) - e.g., 36144
	numericPattern := regexp.MustCompile(`\b(\d{5,6})\b`)
	if matches := numericPattern.FindStringSubmatch(searchText); len(matches) > 0 {
		code := matches[1]
		p.logger.Info("Perplexity code extracted (numeric format)", zap.String("code", code))
		return code
	}

	// Pattern 2: Alphanumeric with hyphen - e.g., aw9s5-y1zoy
	alphanumericPattern := regexp.MustCompile(`\b([a-zA-Z0-9]+-[a-zA-Z0-9]+)\b`)
	if matches := alphanumericPattern.FindStringSubmatch(searchText); len(matches) > 0 {
		code := matches[1]
		p.logger.Info("Perplexity code extracted (alphanumeric format)", zap.String("code", code))
		return code
	}

	// Pattern 3: Fallback to the configured pattern
	if matches := p.codePattern.FindStringSubmatch(searchText); len(matches) > 0 {
		p.logger.Info("Perplexity code extracted (fallback pattern)", zap.String("code", matches[0]))
		return matches[0]
	}

	p.logger.Warn("Code not found after marker in Perplexity email",
		zap.String("searchText_preview", truncateString(searchText, 300)))
	return NotFoundCode
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
	// Si el texto contiene caracteres quoted-printable, intentar decodificar
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
