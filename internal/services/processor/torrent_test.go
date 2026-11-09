package processor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
	"automation-hub/internal/models"
	"automation-hub/internal/services/media"
)

func TestNewTorrentProcessor(t *testing.T) {
	logger := zap.NewNop()
	webhookCfg := &config.WebhookProcessorConfig{
		TelegramChatID:  "123",
		TelegramMessage: "Done %s",
	}

	proc := NewTorrentProcessor(nil, webhookCfg, logger)
	if proc == nil {
		t.Fatal("Expected NewTorrentProcessor to return non-nil instance")
	}
	if proc.notifier.chatID != "123" || proc.notifier.template != "Done %s" {
		t.Errorf("notifier = {chat %q, template %q}, want the hook's chat and template", proc.notifier.chatID, proc.notifier.template)
	}
}

func TestTorrentProcessor_Process(t *testing.T) {
	sender := &fakeSender{}
	webhookCfg := &config.WebhookProcessorConfig{
		TelegramChatID:  "123",
		TelegramMessage: "Download finished: %s at %s",
	}

	proc := NewTorrentProcessor(sender, webhookCfg, zap.NewNop())
	err := proc.Process(models.TorrentNotification{TorrentName: "Dune.2021.1080p", SavePath: "/downloads/movies"})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if len(sender.messages) != 1 || sender.messages[0] != "Download finished: Dune.2021.1080p at /downloads/movies" {
		t.Errorf("sent %q, want the filled template", sender.messages)
	}
	if len(sender.chatIDs) != 1 || sender.chatIDs[0] != "123" {
		t.Errorf("chat IDs = %q, want [123]", sender.chatIDs)
	}
}

type fakeLookup struct {
	metadata *media.Metadata
	err      error
}

func (f fakeLookup) Lookup(context.Context, media.Release) (*media.Metadata, error) {
	return f.metadata, f.err
}

type fakeCards struct {
	photos, texts     int
	photoErr, textErr error
	caption           string
}

func (f *fakeCards) SendPhoto(_, _, caption string) error {
	f.photos++
	f.caption = caption
	return f.photoErr
}
func (f *fakeCards) SendHTML(_, caption string) error {
	f.texts++
	f.caption = caption
	return f.textErr
}
func TestTorrentMetadataFallbacks(t *testing.T) {
	for _, tt := range []struct {
		name                         string
		metadata                     *media.Metadata
		lookupErr, photoErr, textErr error
		photos, texts, templates     int
		fail                         bool
	}{
		{name: "not found", lookupErr: media.ErrNotFound, templates: 1},
		{name: "lookup error", lookupErr: errors.New("failed"), templates: 1},
		{name: "no poster", metadata: &media.Metadata{Title: "Dune"}, texts: 1},
		{name: "photo rejected", metadata: &media.Metadata{Title: "Dune", Poster: "poster"}, photoErr: errors.New("failed"), photos: 1, texts: 1},
		{name: "both rejected", metadata: &media.Metadata{Title: "Dune", Poster: "poster"}, photoErr: errors.New("failed"), textErr: errors.New("failed"), photos: 1, texts: 1, fail: true},
		{name: "full", metadata: &media.Metadata{Title: "Dune", Poster: "poster", IMDbID: "tt123"}, photos: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sender := &fakeSender{}
			cards := &fakeCards{photoErr: tt.photoErr, textErr: tt.textErr}
			cat, err := messages.Load(nil)
			if err != nil {
				t.Fatal(err)
			}
			p := NewTorrentProcessor(sender, &config.WebhookProcessorConfig{TelegramChatID: "123", TelegramMessage: "%s %s"}, zap.NewNop())
			p.ConfigureMetadata(fakeLookup{tt.metadata, tt.lookupErr}, cards, cat)
			err = p.Process(models.TorrentNotification{TorrentName: "Dune.2021.1080p", SavePath: "/secret"})
			if (err != nil) != tt.fail || cards.photos != tt.photos || cards.texts != tt.texts || len(sender.messages) != tt.templates {
				t.Fatalf("error %v photos %d texts %d templates %d", err, cards.photos, cards.texts, len(sender.messages))
			}
			if strings.Contains(cards.caption, "/secret") {
				t.Fatal("save path leaked")
			}
		})
	}
}
