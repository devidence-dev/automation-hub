package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"automation-hub/internal/config"
)

const (
	// maxReportBody caps the multipart body. A Renovate report for about 10
	// repositories is well under this.
	maxReportBody = 10 << 20
	// telegramLimit is Telegram's message size limit. Lengths are counted in
	// bytes of the final HTML, which over-counts and so never exceeds it.
	telegramLimit = 4096
)

// forkTag marks forks in the document rows.
const forkTag = " (fork, solo informativo)" //nolint:misspell // Spanish text fixed by the spec

var htmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

type reportSender interface {
	SendHTML(chatID, text string) error
	SendDocument(chatID, filename, caption string, data []byte) error
}

// DependencyReportHandler receives the Renovate report posted by the
// dependency-report Job and delivers it to Telegram.
type DependencyReportHandler struct {
	cfg    config.DependencyReportConfig
	sender reportSender
	logger *zap.Logger
	now    func() time.Time // when the report was received; a field so tests can set it
}

func NewDependencyReportHandler(cfg config.DependencyReportConfig, sender reportSender, logger *zap.Logger) *DependencyReportHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DependencyReportHandler{cfg: cfg, sender: sender, logger: logger, now: time.Now}
}

// reportDocument is the HTML report, built before the response so a template
// error is logged without losing the caption.
type reportDocument struct {
	name string
	data []byte
}

// ServeHTTP validates the request, answers 202 and sends the report in the
// background, so a slow Telegram cannot hit the server's WriteTimeout.
func (h *DependencyReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Report-Token")), []byte(h.cfg.WebhookToken)) != 1 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxReportBody)
	if err := r.ParseMultipartForm(maxReportBody); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}
		h.logger.Warn("Invalid dependency report request", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	exitCode, err := strconv.Atoi(strings.TrimSpace(r.PostFormValue("exit_code")))
	if err != nil {
		http.Error(w, "Invalid exit_code", http.StatusBadRequest)
		return
	}
	pod := r.PostFormValue("pod")
	logTail, _, err := formFile(r, "log_tail")
	if err != nil {
		http.Error(w, "Invalid log_tail", http.StatusBadRequest)
		return
	}
	reportJSON, hasReport, err := formFile(r, "report")
	if err != nil {
		http.Error(w, "Invalid report", http.StatusBadRequest)
		return
	}

	var text string         // the caption, or the failure message
	var doc *reportDocument // only on the success path
	if exitCode != 0 || !hasReport {
		text = buildFailureMessage(exitCode, pod, string(logTail))
	} else {
		rep, err := parseReport(reportJSON)
		if err != nil {
			h.logger.Warn("Invalid dependency report JSON", zap.Error(err))
			http.Error(w, "Invalid report", http.StatusBadRequest)
			return
		}
		text, doc = h.buildReport(rep)
	}

	w.WriteHeader(http.StatusAccepted)
	h.logger.Info("Dependency report accepted", zap.String("pod", pod), zap.Int("exit_code", exitCode), zap.Bool("document", doc != nil))
	go h.deliver(text, doc)
}

// buildReport renders the caption and the document. A failure is logged: without a
// caption nothing is sent, and without a document the caption goes out as text.
func (h *DependencyReportHandler) buildReport(rep renovateReport) (string, *reportDocument) {
	receivedAt := h.now()
	caption, err := buildCaption(rep, h.cfg.Forks, receivedAt)
	if err != nil {
		h.logger.Error("Failed to build the dependency report caption", zap.Error(err))
		return "", nil
	}
	name, err := documentName(receivedAt)
	if err != nil {
		h.logger.Error("Failed to name the dependency report document", zap.Error(err))
		return caption, nil
	}
	data, err := buildReportDocument(rep, h.cfg.Forks, receivedAt)
	if err != nil {
		h.logger.Error("Failed to build the dependency report document", zap.Error(err))
		return caption, nil
	}
	return caption, &reportDocument{name: name, data: data}
}

// deliver runs after the response. It sends the document with the caption. When that
// fails, the caption goes as a text message, so the summary still arrives. A failure
// message has no document and is sent as text.
func (h *DependencyReportHandler) deliver(text string, doc *reportDocument) {
	if doc != nil {
		err := h.sender.SendDocument(h.cfg.TelegramChatID, doc.name, text, doc.data)
		if err == nil {
			return
		}
		h.logger.Error("Failed to send dependency report document", zap.String("file", doc.name), zap.Error(err))
	}
	if text == "" {
		return
	}
	if err := h.sender.SendHTML(h.cfg.TelegramChatID, text); err != nil {
		h.logger.Error("Failed to send dependency report message", zap.Error(err))
	}
}

// formFile reads an uploaded file part. A missing part is not an error.
func formFile(r *http.Request, name string) ([]byte, bool, error) {
	file, _, err := r.FormFile(name)
	if errors.Is(err, http.ErrMissingFile) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(file)
	return data, true, err
}

// The types below hold only the fields of Renovate's report that the format
// uses (lib/instrumentation/types.ts in Renovate 44.151.0).
type renovateReport struct {
	Problems     []problem             `json:"problems"`
	Repositories map[string]repoReport `json:"repositories"`
}

type repoReport struct {
	Problems     []problem                `json:"problems"`
	PackageFiles map[string][]packageFile `json:"packageFiles"`
}

type problem struct {
	Level int    `json:"level"`
	Msg   string `json:"msg"`
}

type packageFile struct {
	PackageFile string      `json:"packageFile"`
	Deps        []reportDep `json:"deps"`
}

type reportDep struct {
	DepName        *string        `json:"depName"`
	Datasource     *string        `json:"datasource"`
	CurrentVersion *string        `json:"currentVersion"`
	CurrentValue   *string        `json:"currentValue"`
	Updates        []reportUpdate `json:"updates"`
}

type reportUpdate struct {
	UpdateType string  `json:"updateType"`
	NewVersion *string `json:"newVersion"`
	NewValue   *string `json:"newValue"`
}

func parseReport(data []byte) (renovateReport, error) {
	var rep renovateReport
	if err := json.Unmarshal(data, &rep); err != nil {
		return renovateReport{}, fmt.Errorf("parse renovate report: %w", err)
	}
	return rep, nil
}

func firstNonNil(def string, values ...*string) string {
	for _, v := range values {
		if v != nil {
			return *v
		}
	}
	return def
}

// buildFailureMessage reports a failed Renovate run with the last lines of its
// log that fit in one message, HTML-escaped inside <pre>.
func buildFailureMessage(exitCode int, pod, logTail string) string {
	head := fmt.Sprintf("❌ Renovate falló (exit %d) en el pod %s", exitCode, htmlEscaper.Replace(pod))
	const open, closeTag = "\n<pre>", "</pre>"
	budget := telegramLimit - len(head) - len(open) - len(closeTag)

	var kept []string // escaped lines, newest first
	used := 0
	rawLines := strings.Split(strings.TrimRight(logTail, "\r\n"), "\n")
	for i := len(rawLines) - 1; i >= 0; i-- {
		line := htmlEscaper.Replace(rawLines[i])
		cost := len(line)
		if len(kept) > 0 {
			cost++ // joining newline
		}
		if used+cost > budget {
			if len(kept) == 0 {
				kept = append(kept, escapedSuffix(rawLines[i], budget))
			}
			break
		}
		kept = append(kept, line)
		used += cost
	}
	if len(kept) == 0 || (len(kept) == 1 && kept[0] == "") {
		return head
	}
	for i, j := 0, len(kept)-1; i < j; i, j = i+1, j-1 {
		kept[i], kept[j] = kept[j], kept[i]
	}
	return head + open + strings.Join(kept, "\n") + closeTag
}

// escapedSuffix keeps the tail of one over-long line, escaping rune by rune so
// the cut never lands inside an entity.
func escapedSuffix(line string, budget int) string {
	runes := []rune(line)
	used := 0
	start := len(runes)
	for start > 0 {
		cost := len(htmlEscaper.Replace(string(runes[start-1])))
		if used+cost > budget {
			break
		}
		used += cost
		start--
	}
	return htmlEscaper.Replace(string(runes[start:]))
}
