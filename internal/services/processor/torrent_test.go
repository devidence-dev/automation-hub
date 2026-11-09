package processor

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
	"automation-hub/internal/models"
	"automation-hub/internal/services/media"
	"automation-hub/internal/services/telegram"
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

// fakeCards records every card send in order. A photo whose URL is in rejected fails, and textErr fails the text send.
type fakeCards struct {
	rejected  []string
	textErr   error
	photoURLs []string
	texts     int
	caption   string
	buttons   [][]telegram.LinkButton // the buttons of every send, photo and text alike
}

func (f *fakeCards) SendPhoto(_, photoURL, caption string, buttons ...telegram.LinkButton) error {
	f.photoURLs = append(f.photoURLs, photoURL)
	f.record(caption, buttons)
	if slices.Contains(f.rejected, photoURL) {
		return errors.New("photo rejected")
	}
	return nil
}

func (f *fakeCards) SendHTMLButtons(_, caption string, buttons ...telegram.LinkButton) error {
	f.texts++
	f.record(caption, buttons)
	return f.textErr
}

func (f *fakeCards) record(caption string, buttons []telegram.LinkButton) {
	f.caption = caption
	f.buttons = append(f.buttons, buttons)
}

// metadataCase is one row of TestTorrentMetadataFallbacks: the lookup, the sends Telegram rejects and the expected sends.
type metadataCase struct {
	name      string
	metadata  *media.Metadata
	lookupErr error
	rejected  []string
	textErr   error
	photos    []string
	texts     int
	templates int
	button    bool
	fail      bool
}

func TestTorrentMetadataFallbacks(t *testing.T) {
	for _, tt := range []metadataCase{
		{name: "not found", lookupErr: media.ErrNotFound, templates: 1},
		{name: "lookup error", lookupErr: errors.New("failed"), templates: 1},
		{name: "no poster", metadata: &media.Metadata{Title: "Dune"}, texts: 1},
		{name: "no poster with IMDb id", metadata: &media.Metadata{Title: "Dune", IMDbID: "tt123"}, texts: 1, button: true},
		{name: "resized poster rejected", metadata: &media.Metadata{Title: "Dune", Poster: "resized", PosterFallback: "original", IMDbID: "tt123"}, rejected: []string{"resized"}, photos: []string{"resized", "original"}, button: true},
		{name: "all posters rejected", metadata: &media.Metadata{Title: "Dune", Poster: "resized", PosterFallback: "original", IMDbID: "tt123"}, rejected: []string{"resized", "original"}, photos: []string{"resized", "original"}, texts: 1, button: true},
		{name: "text fails", metadata: &media.Metadata{Title: "Dune"}, textErr: errors.New("failed"), texts: 1, fail: true},
		{name: "both sends fail", metadata: &media.Metadata{Title: "Dune", Poster: "poster"}, rejected: []string{"poster"}, textErr: errors.New("failed"), photos: []string{"poster"}, texts: 1, fail: true},
		{name: "full", metadata: &media.Metadata{Title: "Dune", Poster: "poster", IMDbID: "tt123"}, photos: []string{"poster"}, button: true},
		{name: "full without IMDb id", metadata: &media.Metadata{Title: "Dune", Poster: "poster"}, photos: []string{"poster"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			runMetadataCase(t, tt)
		})
	}
}

// runMetadataCase sends a torrent notification through a processor with the case's lookup and Telegram fakes,
// and checks the sends against the case.
func runMetadataCase(t *testing.T, tt metadataCase) {
	t.Helper()
	sender := &fakeSender{}
	cards := &fakeCards{rejected: tt.rejected, textErr: tt.textErr}
	cat, err := messages.Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	p := NewTorrentProcessor(sender, &config.WebhookProcessorConfig{TelegramChatID: "123", TelegramMessage: "%s %s"}, zap.NewNop())
	p.ConfigureMetadata(fakeLookup{tt.metadata, tt.lookupErr}, cards, cat, config.DefaultMetadataPlotMaxChars)
	err = p.Process(models.TorrentNotification{TorrentName: "Dune.2021.1080p", SavePath: "/secret"})
	if (err != nil) != tt.fail || !slices.Equal(cards.photoURLs, tt.photos) || cards.texts != tt.texts || len(sender.messages) != tt.templates {
		t.Fatalf("error %v photos %q texts %d templates %d", err, cards.photoURLs, cards.texts, len(sender.messages))
	}
	checkButtons(t, cards.buttons, tt.button)
	if strings.Contains(cards.caption, "/secret") {
		t.Fatal("save path leaked")
	}
}

// checkButtons asserts that every send carried the IMDb button when want is true, and no button otherwise.
func checkButtons(t *testing.T, sent [][]telegram.LinkButton, want bool) {
	t.Helper()
	wantButtons := []telegram.LinkButton{{Text: "Open on IMDb", URL: "https://www.imdb.com/title/tt123/"}}
	for _, buttons := range sent {
		if !want && len(buttons) != 0 {
			t.Fatalf("buttons %v, want none", buttons)
		}
		if want && !slices.Equal(buttons, wantButtons) {
			t.Fatalf("buttons %v, want %v", buttons, wantButtons)
		}
	}
}

func TestPosterURLs(t *testing.T) {
	for _, tt := range []struct {
		name string
		m    media.Metadata
		want []string
	}{
		{name: "resized and original", m: media.Metadata{Poster: "resized", PosterFallback: "original"}, want: []string{"resized", "original"}},
		{name: "poster only", m: media.Metadata{Poster: "poster"}, want: []string{"poster"}},
		{name: "fallback only", m: media.Metadata{PosterFallback: "original"}, want: []string{"original"}},
		{name: "none", m: media.Metadata{}, want: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := posterURLs(&tt.m); !slices.Equal(got, tt.want) {
				t.Fatalf("posterURLs() = %q, want %q", got, tt.want)
			}
		})
	}
}
