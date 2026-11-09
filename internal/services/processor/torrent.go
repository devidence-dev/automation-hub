package processor

import (
	"context"
	"errors"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
	"automation-hub/internal/models"
	"automation-hub/internal/services/media"
	"go.uber.org/zap"
)

type cardSender interface {
	SendPhoto(chatID, photoURL, caption string) error
	SendHTML(chatID, message string) error
}
type metadataLookup interface {
	Lookup(context.Context, media.Release) (*media.Metadata, error)
}

// TorrentProcessor sends the completion message through the hook's template or a metadata card.
type TorrentProcessor struct {
	notifier TemplateNotifier
	logger   *zap.Logger
	lookup   metadataLookup
	cards    cardSender
	catalog  *messages.Catalog
}

func NewTorrentProcessor(sender messageSender, webhookConfig *config.WebhookProcessorConfig, logger *zap.Logger) *TorrentProcessor {
	return &TorrentProcessor{notifier: TemplateNotifier{sender: sender, chatID: webhookConfig.TelegramChatID, template: webhookConfig.TelegramMessage}, logger: logger}
}

// ConfigureMetadata enables enriched cards for this processor.
func (p *TorrentProcessor) ConfigureMetadata(lookup media.Provider, sender cardSender, catalog *messages.Catalog) {
	p.lookup = lookup
	p.cards = sender
	p.catalog = catalog
}
func (p *TorrentProcessor) Process(notification models.TorrentNotification) error {
	p.logger.Debug("Sending torrent notification", zap.String("torrent", notification.TorrentName))
	if p.lookup != nil && p.cards != nil && p.catalog != nil {
		if handled, err := p.sendCard(notification.TorrentName); handled {
			return err
		}
	}
	return p.notifier.Notify(notification.TorrentName, notification.SavePath)
}

// sendCard sends the metadata card. handled=false means no card was attempted and the template should be sent.
func (p *TorrentProcessor) sendCard(torrentName string) (bool, error) {
	release := media.ParseRelease(torrentName)
	if release.Title == "" {
		return false, nil
	}
	metadata := p.lookupMetadata(release)
	if metadata == nil {
		return false, nil
	}
	caption, err := (CardPresenter{Catalog: p.catalog}).Present(metadata, release)
	if err != nil {
		p.logger.Warn("Metadata card failed", zap.Error(err))
		return false, nil
	}
	if metadata.Poster != "" && p.cards.SendPhoto(p.notifier.chatID, metadata.Poster, caption) == nil {
		return true, nil
	}
	return true, p.cards.SendHTML(p.notifier.chatID, caption)
}

// lookupMetadata returns the metadata for release, or nil when none is available. Lookup failures are logged here.
func (p *TorrentProcessor) lookupMetadata(release media.Release) *media.Metadata {
	metadata, err := p.lookup.Lookup(context.Background(), release)
	if err == nil {
		return metadata
	}
	if errors.Is(err, media.ErrNotFound) {
		p.logger.Debug("Metadata not found")
	} else {
		p.logger.Warn("Metadata lookup failed", zap.Error(err))
	}
	return nil
}
