package processor

import (
	"context"
	"errors"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
	"automation-hub/internal/models"
	"automation-hub/internal/services/media"
	"automation-hub/internal/services/telegram"
	"go.uber.org/zap"
)

type cardSender interface {
	SendPhoto(chatID, photoURL, caption string, buttons ...telegram.LinkButton) error
	SendHTMLButtons(chatID, message string, buttons ...telegram.LinkButton) error
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
	// plotMaxChars is the plot length limit of the card, 0 hides the plot.
	plotMaxChars int
}

func NewTorrentProcessor(sender messageSender, webhookConfig *config.WebhookProcessorConfig, logger *zap.Logger) *TorrentProcessor {
	return &TorrentProcessor{notifier: TemplateNotifier{sender: sender, chatID: webhookConfig.TelegramChatID, template: webhookConfig.TelegramMessage}, logger: logger}
}

// ConfigureMetadata enables enriched cards for this processor. plotMaxChars is the effective plot length limit.
func (p *TorrentProcessor) ConfigureMetadata(lookup media.Provider, sender cardSender, catalog *messages.Catalog, plotMaxChars int) {
	p.lookup = lookup
	p.cards = sender
	p.catalog = catalog
	p.plotMaxChars = plotMaxChars
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
	rendered, buttons, err := p.renderCard(metadata, release)
	if err != nil {
		p.logger.Warn("Metadata card failed", zap.Error(err))
		return false, nil
	}
	if p.sendPhoto(metadata, rendered.Caption, buttons) {
		return true, nil
	}
	return true, p.cards.SendHTMLButtons(p.notifier.chatID, rendered.Caption, buttons...)
}

// renderCard renders the card of metadata and the button that opens its IMDb page, none without an IMDb id.
func (p *TorrentProcessor) renderCard(metadata *media.Metadata, release media.Release) (Card, []telegram.LinkButton, error) {
	rendered, err := (CardPresenter{Catalog: p.catalog, PlotMaxChars: p.plotMaxChars}).Present(metadata, release)
	if err != nil {
		return Card{}, nil, err
	}
	if rendered.URL == "" {
		return rendered, nil, nil
	}
	label, err := p.catalog.Render("torrent.card.button", nil)
	if err != nil {
		return Card{}, nil, err
	}
	return rendered, []telegram.LinkButton{{Text: label, URL: rendered.URL}}, nil
}

// sendPhoto sends the caption with the first poster URL that Telegram accepts, and reports whether one was sent.
func (p *TorrentProcessor) sendPhoto(metadata *media.Metadata, caption string, buttons []telegram.LinkButton) bool {
	for _, photoURL := range posterURLs(metadata) {
		if p.cards.SendPhoto(p.notifier.chatID, photoURL, caption, buttons...) == nil {
			return true
		}
	}
	return false
}

// posterURLs returns the poster URLs to try in order: the resized one, then the original. Empty URLs are left out.
func posterURLs(metadata *media.Metadata) []string {
	var urls []string
	for _, u := range []string{metadata.Poster, metadata.PosterFallback} {
		if u != "" {
			urls = append(urls, u)
		}
	}
	return urls
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
