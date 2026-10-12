package depreport

import (
	"slices"
	"strconv"
	"strings"

	"automation-hub/internal/messages"
)

// telegramLimit is Telegram's message size limit. Lengths are counted in bytes of the final
// HTML, which over-counts and so never exceeds it.
const telegramLimit = 4096

// BuildFailure returns the message for a failed Renovate run: the catalog title with the exit code
// and the escaped pod, then the last lines of the log that fit in one message, escaped inside <pre>.
func BuildFailure(cat *messages.Catalog, exitCode int, pod, logTail string) (string, error) {
	r := &renderer{cat: cat}
	head := r.text(KeyFailureTitle, map[string]string{
		"exit": strconv.Itoa(exitCode),
		"pod":  htmlEscaper.Replace(pod),
	})
	if r.err != nil {
		return "", r.err
	}
	return logTailBlock(head, logTail), nil
}

// logTailBlock appends the escaped tail of the log to head, within the message limit.
func logTailBlock(head, logTail string) string {
	const open, closeTag = "\n<pre>", "</pre>"
	kept := keptLines(logTail, telegramLimit-len(head)-len(open)-len(closeTag))
	if len(kept) == 0 || (len(kept) == 1 && kept[0] == "") {
		return head
	}
	return head + open + strings.Join(kept, "\n") + closeTag
}

// keptLines returns the escaped last lines of logTail that fit in budget bytes, oldest first.
func keptLines(logTail string, budget int) []string {
	var kept []string // newest first
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
	slices.Reverse(kept)
	return kept
}

// escapedSuffix keeps the tail of one over-long line, escaping rune by rune so the cut never
// lands inside an entity.
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
