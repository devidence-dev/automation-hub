package depreport

import (
	"go.uber.org/zap"
)

// ReportSender is what the dependency report needs from the notifications bot. It is the only
// definition of this interface.
type ReportSender interface {
	SendHTML(chatID, text string) error
	SendDocument(chatID, filename, caption string, data []byte) error
}

// Delivery is one report to send: the caption, or the failure text, and on success the document.
type Delivery struct {
	Text     string
	Filename string
	Document []byte // nil when there is no document
}

// Deliverer sends deliveries to one chat.
type Deliverer struct {
	sender ReportSender
	chatID string
	logger *zap.Logger
}

// NewDeliverer returns a deliverer for chatID.
func NewDeliverer(sender ReportSender, chatID string, logger *zap.Logger) *Deliverer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Deliverer{sender: sender, chatID: chatID, logger: logger}
}

// Deliver sends the document with the caption. When the document fails, the caption goes as a text
// message, so the summary still arrives. A failure message has no document and goes as text.
// Errors are logged: the request was answered before the send.
func (d *Deliverer) Deliver(del Delivery) {
	if del.Document != nil {
		err := d.sender.SendDocument(d.chatID, del.Filename, del.Text, del.Document)
		if err == nil {
			return
		}
		d.logger.Error("Failed to send dependency report document", zap.String("file", del.Filename), zap.Error(err))
	}
	if del.Text == "" {
		return
	}
	if err := d.sender.SendHTML(d.chatID, del.Text); err != nil {
		d.logger.Error("Failed to send dependency report message", zap.Error(err))
	}
}
