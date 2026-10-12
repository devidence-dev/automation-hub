package email

import (
	"context"
	"time"

	"go.uber.org/zap"

	"automation-hub/internal/models"
)

// MailSource is the transport: it offers each unread email from the senders to handle, and sets the
// \Seen flag of the emails for which handle returns true.
type MailSource interface {
	Unread(senders []string, handle func(models.Email) bool) error
}

// Dispatcher is what the poller needs from the processors: the senders to search and the dispatch of each email.
type Dispatcher interface {
	Senders() []string
	Dispatch(email models.Email) (models.EmailProcessor, error)
}

// Poller runs one poll per interval and decides, for each email, whether it is marked as read.
type Poller struct {
	source     MailSource
	dispatcher Dispatcher
	interval   time.Duration
	logger     *zap.Logger
}

func NewPoller(source MailSource, dispatcher Dispatcher, interval time.Duration, logger *zap.Logger) *Poller {
	return &Poller{
		source:     source,
		dispatcher: dispatcher,
		interval:   interval,
		logger:     logger,
	}
}

// Run polls until ctx is done. A failed poll is logged and the next one retries.
func (p *Poller) Run(ctx context.Context) {
	p.logger.Info("Starting email monitoring", zap.Duration("polling_interval", p.interval))

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll()
		}
	}
}

func (p *Poller) poll() {
	if err := p.source.Unread(p.dispatcher.Senders(), p.handle); err != nil {
		p.logger.Error("Email poll failed", zap.Error(err))
	}
}

// handle dispatches one email. It returns true only when a processor matched, succeeded, and its
// ShouldMarkAsRead returns true. Every other outcome leaves the email unread.
func (p *Poller) handle(email models.Email) bool {
	processor, err := p.dispatcher.Dispatch(email)
	if err != nil || processor == nil {
		return false
	}
	marker, ok := processor.(models.ReadMarker)
	if !ok || !marker.ShouldMarkAsRead() {
		p.logger.Info("Email processed but NOT marked as read",
			zap.String("from", email.From),
			zap.String("subject", email.Subject))
		return false
	}
	p.logger.Info("Marking email as read",
		zap.String("from", email.From),
		zap.String("subject", email.Subject))
	return true
}
