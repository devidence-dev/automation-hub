package handlers

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
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
	// maxDepLines caps the dependency lines per repository message.
	maxDepLines = 40
	// sendTimeout bounds the background delivery of one report.
	sendTimeout = 2 * time.Minute
)

// forkTag marks forks in the report header.
const forkTag = " (fork, solo informativo)" //nolint:misspell // Spanish text fixed by the spec

var htmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

type htmlSender interface {
	SendHTML(chatID, text string) error
}

// DependencyReportHandler receives the Renovate report posted by the
// dependency-report Job and delivers it to Telegram.
type DependencyReportHandler struct {
	cfg    config.DependencyReportConfig
	sender htmlSender
	logger *zap.Logger
}

func NewDependencyReportHandler(cfg config.DependencyReportConfig, sender htmlSender, logger *zap.Logger) *DependencyReportHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DependencyReportHandler{cfg: cfg, sender: sender, logger: logger}
}

// ServeHTTP validates the request, answers 202 and sends the messages in the
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

	var messages []string
	if exitCode != 0 || !hasReport {
		messages = []string{buildFailureMessage(exitCode, pod, string(logTail))}
	} else {
		rep, err := parseReport(reportJSON)
		if err != nil {
			h.logger.Warn("Invalid dependency report JSON", zap.Error(err))
			http.Error(w, "Invalid report", http.StatusBadRequest)
			return
		}
		messages = buildMessages(rep, h.cfg.Forks)
	}

	w.WriteHeader(http.StatusAccepted)
	h.logger.Info("Dependency report accepted", zap.String("pod", pod), zap.Int("exit_code", exitCode), zap.Int("messages", len(messages)))
	go h.deliver(messages)
}

// deliver runs after the response, so it uses its own context instead of the request's.
func (h *DependencyReportHandler) deliver(messages []string) {
	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()
	for i, msg := range messages {
		if err := ctx.Err(); err != nil {
			h.logger.Error("Dependency report delivery timed out", zap.Int("sent", i), zap.Int("total", len(messages)), zap.Error(err))
			return
		}
		if err := h.sender.SendHTML(h.cfg.TelegramChatID, msg); err != nil {
			h.logger.Error("Failed to send dependency report message", zap.Int("message", i+1), zap.Int("total", len(messages)), zap.Error(err))
		}
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
	Repositories map[string]repoReport `json:"repositories"`
}

type repoReport struct {
	Problems     []struct{ Level int } `json:"problems"`
	PackageFiles map[string][]struct {
		Deps []reportDep `json:"deps"`
	} `json:"packageFiles"`
	LibYearsWithStatus *struct {
		LibYears         struct{ Total float64 } `json:"libYears"`
		DependencyStatus struct{ Outdated, Total int }
	} `json:"libYearsWithStatus"`
}

type reportDep struct {
	DepName        *string `json:"depName"`
	CurrentVersion *string `json:"currentVersion"`
	CurrentValue   *string `json:"currentValue"`
	Updates        []struct {
		UpdateType string  `json:"updateType"`
		NewVersion *string `json:"newVersion"`
		NewValue   *string `json:"newValue"`
	} `json:"updates"`
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

// depLine formats one dependency as HTML, or returns "" when it has no
// major, minor or patch update (digest, pin and the rest are ignored).
func depLine(dep reportDep) string {
	var updates []string
	emoji := "🟡"
	for _, u := range dep.Updates {
		switch u.UpdateType {
		case "major":
			emoji = "🔴"
		case "minor", "patch":
		default:
			continue
		}
		updates = append(updates, fmt.Sprintf("%s (%s)", htmlEscaper.Replace(firstNonNil("?", u.NewVersion, u.NewValue)), u.UpdateType))
	}
	if len(updates) == 0 {
		return ""
	}
	return fmt.Sprintf("%s <code>%s</code> %s → %s", emoji,
		htmlEscaper.Replace(firstNonNil("?", dep.DepName)),
		htmlEscaper.Replace(firstNonNil("?", dep.CurrentVersion, dep.CurrentValue)),
		strings.Join(updates, ", "))
}

// buildMessages returns one HTML message per repository with pending updates or
// scan errors, or the single "all up to date" message. It follows the interim
// jq in the workflows repo: sorted repos, sorted and de-duplicated lines.
func buildMessages(rep renovateReport, forks []string) []string {
	repos := make([]string, 0, len(rep.Repositories))
	for name := range rep.Repositories {
		repos = append(repos, name)
	}
	sort.Strings(repos)

	var messages []string
	for _, name := range repos {
		if msg := buildRepoMessage(name, rep.Repositories[name], forks); msg != "" {
			messages = append(messages, msg)
		}
	}
	if len(messages) == 0 {
		return []string{"✅ Dependency report: todo al día."}
	}
	return messages
}

func buildRepoMessage(name string, repo repoReport, forks []string) string {
	seen := map[string]struct{}{}
	var lines []string
	for _, files := range repo.PackageFiles {
		for _, file := range files {
			for _, dep := range file.Deps {
				line := depLine(dep)
				if _, dup := seen[line]; line == "" || dup {
					continue
				}
				seen[line] = struct{}{}
				lines = append(lines, line)
			}
		}
	}
	sort.Strings(lines) // like jq's unique; also puts 🔴 before 🟡

	errCount := 0
	for _, p := range repo.Problems {
		if p.Level >= 50 {
			errCount++
		}
	}
	if len(lines) == 0 && errCount == 0 {
		return ""
	}

	head := "<b>📦 " + htmlEscaper.Replace(name) + "</b>"
	for _, fork := range forks {
		if fork == name {
			head += forkTag
			break
		}
	}
	headLines := []string{head}
	if ly := repo.LibYearsWithStatus; ly != nil {
		headLines = append(headLines, fmt.Sprintf("%d/%d desactualizadas · %s libyears",
			ly.DependencyStatus.Outdated, ly.DependencyStatus.Total,
			strconv.FormatFloat(math.Round(ly.LibYears.Total*10)/10, 'f', -1, 64)))
	}
	if errCount > 0 {
		headLines = append(headLines, fmt.Sprintf("⚠️ %d errores en el scan", errCount))
	}

	// Cut by whole lines (never inside a tag or entity): start at the 40-line cap
	// and drop lines until the message fits Telegram's limit.
	shown := min(len(lines), maxDepLines)
	for ; ; shown-- {
		out := append([]string(nil), headLines...)
		if len(lines) > 0 {
			out = append(out, "")
			out = append(out, lines[:shown]...)
			if shown < len(lines) {
				out = append(out, fmt.Sprintf("… y %d más", len(lines)-shown))
			}
		}
		if msg := strings.Join(out, "\n"); len(msg) <= telegramLimit || shown <= 0 {
			return msg
		}
	}
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
