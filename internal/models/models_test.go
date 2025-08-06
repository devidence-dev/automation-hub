package models

import (
	"testing"
)

func TestEmail(t *testing.T) {
	email := Email{
		From:      "test@example.com",
		Subject:   "Test Subject",
		TextPlain: "Test Body",
		ID:        "test-123",
	}

	if email.From != "test@example.com" {
		t.Errorf("Expected From to be 'test@example.com', got %s", email.From)
	}

	if email.Subject != "Test Subject" {
		t.Errorf("Expected Subject to be 'Test Subject', got %s", email.Subject)
	}

	if email.TextPlain != "Test Body" {
		t.Errorf("Expected TextPlain to be 'Test Body', got %s", email.TextPlain)
	}
}

func TestTorrentNotification(t *testing.T) {
	notification := TorrentNotification{
		TorrentName: "test.torrent",
		SavePath:    "/downloads/test",
	}

	if notification.TorrentName != "test.torrent" {
		t.Errorf("Expected TorrentName to be 'test.torrent', got %s", notification.TorrentName)
	}
}
