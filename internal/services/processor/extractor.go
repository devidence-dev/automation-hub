package processor

import (
	"fmt"
	"regexp"

	"go.uber.org/zap"

	"automation-hub/internal/config"
)

// CodeExtractor is the strategy that finds the code in a prepared email body.
type CodeExtractor interface {
	Extract(text string) (code string, found bool)
}

// newCodeExtractor returns the strategy named by the service's extractor. Its regexes are compiled here, once.
func newCodeExtractor(cfg config.ServiceProcessorConfig, logPreviewLength int, logger *zap.Logger) (CodeExtractor, error) {
	switch cfg.EffectiveExtractor() {
	case config.ExtractorPattern:
		pattern, err := regexp.Compile(cfg.EffectiveCodePattern())
		if err != nil {
			return nil, fmt.Errorf("code_pattern: %w", err)
		}
		return &patternExtractor{pattern: pattern, previewLength: logPreviewLength, logger: logger}, nil
	case config.ExtractorMarker:
		return newMarkerExtractor(cfg, logPreviewLength, logger)
	default:
		return nil, fmt.Errorf("unknown extractor %q", cfg.Extractor)
	}
}

// patternExtractor takes the first match of code_pattern.
type patternExtractor struct {
	pattern       *regexp.Regexp
	previewLength int
	logger        *zap.Logger
}

func (e *patternExtractor) Extract(text string) (string, bool) {
	code, found := matchCode(e.pattern, text)
	reportResult(e.logger, code, found, text, e.previewLength)
	return code, found
}

// markerExtractor finds the first marker in configured order, then tries the candidates on the text
// after it, then falls back to code_pattern when that is set.
type markerExtractor struct {
	markers       []*regexp.Regexp // case-insensitive, in configured order
	candidates    []*regexp.Regexp
	fallback      *regexp.Regexp // nil when code_pattern is unset
	previewLength int
	logger        *zap.Logger
}

func newMarkerExtractor(cfg config.ServiceProcessorConfig, logPreviewLength int, logger *zap.Logger) (*markerExtractor, error) {
	markerExpressions := make([]string, 0, len(cfg.CodeMarkers))
	for _, marker := range cfg.CodeMarkers {
		markerExpressions = append(markerExpressions, "(?i)"+regexp.QuoteMeta(marker))
	}
	markers, err := compileAll(markerExpressions)
	if err != nil {
		return nil, fmt.Errorf("code_markers: %w", err)
	}
	candidates, err := compileAll(cfg.CodeCandidates)
	if err != nil {
		return nil, fmt.Errorf("code_candidates: %w", err)
	}
	var fallback *regexp.Regexp
	if cfg.CodePattern != "" {
		if fallback, err = regexp.Compile(cfg.CodePattern); err != nil {
			return nil, fmt.Errorf("code_pattern: %w", err)
		}
	}
	return &markerExtractor{
		markers:       markers,
		candidates:    candidates,
		fallback:      fallback,
		previewLength: logPreviewLength,
		logger:        logger,
	}, nil
}

func (e *markerExtractor) Extract(text string) (string, bool) {
	code, found := e.find(text)
	reportResult(e.logger, code, found, text, e.previewLength)
	return code, found
}

func (e *markerExtractor) find(text string) (string, bool) {
	after, ok := textAfterMarker(text, e.markers)
	if !ok {
		return "", false
	}
	for _, candidate := range e.candidates {
		if code, found := matchCode(candidate, after); found {
			return code, true
		}
	}
	if e.fallback == nil {
		return "", false
	}
	return matchCode(e.fallback, after)
}

// textAfterMarker returns the text after the first marker, in marker order, that occurs in text.
func textAfterMarker(text string, markers []*regexp.Regexp) (string, bool) {
	for _, marker := range markers {
		if loc := marker.FindStringIndex(text); loc != nil {
			return text[loc[1]:], true
		}
	}
	return "", false
}

// matchCode returns the first match of pattern. Capture group 1 is the code when it matched a non-empty
// value, so a pattern can anchor on a character it must not return, such as "[^#]\b(\d{6})\b".
func matchCode(pattern *regexp.Regexp, text string) (string, bool) {
	matches := pattern.FindStringSubmatch(text)
	if len(matches) == 0 {
		return "", false
	}
	if len(matches) > 1 && matches[1] != "" {
		return matches[1], true
	}
	return matches[0], true
}

// reportResult logs the outcome of an extraction. A miss logs at most previewLength characters of the text.
func reportResult(logger *zap.Logger, code string, found bool, text string, previewLength int) {
	if found {
		logger.Info("Code extracted successfully", zap.String("code", code))
		return
	}
	logger.Warn("Code not found in email", zap.String("text_preview", truncateString(text, previewLength)))
}

func compileAll(expressions []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(expressions))
	for _, expression := range expressions {
		re, err := regexp.Compile(expression)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", expression, err)
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}
