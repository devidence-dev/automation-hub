package processor

import (
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/models"
	"automation-hub/internal/services/telegram"
)

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
			"perplexity": regexp.MustCompile(`\b[a-zA-Z0-9]{5}-[a-zA-Z0-9]{5}\b`),
			"default":    regexp.MustCompile(`\b[a-zA-Z0-9]{4,8}\b`), // patrón genérico
		},
	}

	// Si hay un patrón personalizado, usarlo
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

	// Si no hay patrón personalizado, usar el patrón por defecto del servicio
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
	// Verificar el remitente
	if !strings.Contains(email.From, p.config.EmailFrom) {
		return false
	}

	// Verificar al menos uno de los subjects
	for _, subject := range p.config.EmailSubject {
		if strings.Contains(email.Subject, subject) {
			return true
		}
	}

	return false
}

func (p *GenericEmailProcessor) Process(email models.Email) error {
	// Extraer el código usando el patrón configurado
	code := p.extractCode(email.TextPlain)

	// Formatear el mensaje
	message := fmt.Sprintf(p.config.TelegramMessage, code)

	// Enviar mensaje a Telegram
	return p.telegram.SendMessage(p.config.TelegramChatID, message)
}

func (p *GenericEmailProcessor) extractCode(text string) string {
	matches := p.codePattern.FindStringSubmatch(text)
	if len(matches) > 0 {
		return matches[0]
	}
	return "No encontrado"
}

func (p *GenericEmailProcessor) GetName() string {
	return p.name
}
