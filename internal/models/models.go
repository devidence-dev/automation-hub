package models

// Email represents an email message
type Email struct {
	Subject   string
	From      string
	TextPlain string
	ID        string
}

// TorrentNotification represents a torrent completion notification
type TorrentNotification struct {
	TorrentName string `json:"torrent_name"`
	SavePath    string `json:"save_path"`
}

// EmailProcessor Processor interface for email processors
type EmailProcessor interface {
	ShouldProcess(email Email) bool
	Process(email Email) error
	GetSender() string
}

// ReadMarker is an optional interface of an EmailProcessor: ShouldMarkAsRead reports whether a
// successfully processed email gets the \Seen flag.
type ReadMarker interface {
	ShouldMarkAsRead() bool
}
