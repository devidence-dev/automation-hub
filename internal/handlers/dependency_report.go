package handlers

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
	"automation-hub/internal/services/depreport"
)

// DependencyReportHandler receives the Renovate report posted by the dependency-report Job and
// delivers it to Telegram. It only checks, parses and orchestrates; the report logic is in depreport.
type DependencyReportHandler struct {
	cfg       config.DependencyReportConfig
	cat       *messages.Catalog
	loc       *time.Location
	aggregate func(depreport.Report, []string, time.Time) depreport.Summary
	caption   depreport.Presenter
	document  depreport.Presenter
	deliverer *depreport.Deliverer
	logger    *zap.Logger
	now       func() time.Time // when the report was received; a field so tests can set it
}

// NewDependencyReportHandler returns the handler for cfg. The catalog and the zone come from the
// config; sender is the notifications bot.
func NewDependencyReportHandler(cfg config.DependencyReportConfig, cat *messages.Catalog, loc *time.Location, sender depreport.ReportSender, logger *zap.Logger) *DependencyReportHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DependencyReportHandler{
		cfg:       cfg,
		cat:       cat,
		loc:       loc,
		aggregate: depreport.Aggregate,
		caption:   depreport.NewCaptionPresenter(cat, loc, cfg.DateLayout),
		document:  depreport.NewDocumentPresenter(cat, loc),
		deliverer: depreport.NewDeliverer(sender, cfg.TelegramChatID, logger),
		logger:    logger,
		now:       time.Now,
	}
}

// ServeHTTP checks the token, reads the submission and answers 202 at once. The Telegram messages
// are sent in the background, so a slow Telegram cannot hit the server's WriteTimeout.
func (h *DependencyReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	sub, err := depreport.ParseSubmission(r, h.cfg.MaxBodyBytes)
	if err != nil {
		h.rejectSubmission(w, err)
		return
	}
	if sub.Failed() {
		h.answer(w, sub, h.failureDelivery(sub))
		return
	}
	rep, err := depreport.ParseReport(sub.Report)
	if err != nil {
		h.logger.Warn("Invalid dependency report JSON", zap.Error(err))
		http.Error(w, "Invalid report", http.StatusBadRequest)
		return
	}
	h.answer(w, sub, h.reportDelivery(rep))
}

// authorized compares the configured token header to the token in constant time.
func (h *DependencyReportHandler) authorized(r *http.Request) bool {
	got := []byte(r.Header.Get(h.cfg.TokenHeader))
	return subtle.ConstantTimeCompare(got, []byte(h.cfg.WebhookToken)) == 1
}

// rejectSubmission answers a request that could not be read: 413 when it is too large, else 400.
func (h *DependencyReportHandler) rejectSubmission(w http.ResponseWriter, err error) {
	if errors.Is(err, depreport.ErrTooLarge) {
		http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
		return
	}
	h.logger.Warn("Invalid dependency report request", zap.Error(err))
	http.Error(w, "Invalid request", http.StatusBadRequest)
}

// answer sends 202 and then delivers in the background.
func (h *DependencyReportHandler) answer(w http.ResponseWriter, sub depreport.Submission, delivery depreport.Delivery) {
	w.WriteHeader(http.StatusAccepted)
	h.logger.Info("Dependency report accepted", zap.String("pod", sub.Pod), zap.Int("exit_code", sub.ExitCode), zap.Bool("document", delivery.Document != nil))
	go h.deliverer.Deliver(delivery)
}

// failureDelivery is the failure message of a Renovate run that failed. It has no document.
func (h *DependencyReportHandler) failureDelivery(sub depreport.Submission) depreport.Delivery {
	text, err := depreport.BuildFailure(h.cat, sub.ExitCode, sub.Pod, string(sub.LogTail))
	if err != nil {
		h.logger.Error("Failed to build the dependency failure message", zap.Error(err))
		return depreport.Delivery{}
	}
	return depreport.Delivery{Text: text}
}

// reportDelivery aggregates the report once and presents it. Without a caption nothing is sent;
// without a document the caption goes out as text.
func (h *DependencyReportHandler) reportDelivery(rep depreport.Report) depreport.Delivery {
	receivedAt := h.now()
	sum := h.aggregate(rep, h.cfg.Forks, receivedAt)

	caption, err := h.caption.Present(sum)
	if err != nil {
		h.logger.Error("Failed to build the dependency report caption", zap.Error(err))
		return depreport.Delivery{}
	}
	doc, err := h.document.Present(sum)
	if err != nil {
		h.logger.Error("Failed to build the dependency report document", zap.Error(err))
		return depreport.Delivery{Text: string(caption)}
	}
	return depreport.Delivery{
		Text:     string(caption),
		Filename: depreport.DocumentName(receivedAt, h.loc, h.cfg.DocumentName, h.cfg.DateLayout),
		Document: doc,
	}
}
