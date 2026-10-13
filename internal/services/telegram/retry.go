package telegram

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// sender is the part of *tgbotapi.BotAPI that the retry decorator wraps.
type sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

// Delayer decides whether a failed attempt is retried and how long to wait before the next one.
// attempt is the 1-based number of the attempt that failed.
type Delayer interface {
	Delay(attempt int, err error) (time.Duration, bool)
}

// ExponentialBackoff waits Base after the first failed attempt and doubles the wait after each
// further one, never longer than Ceiling. A 429 answer with retry_after waits that many seconds
// instead, and is not retried when that wait exceeds Ceiling.
type ExponentialBackoff struct {
	Base    time.Duration
	Ceiling time.Duration
}

// Delay returns the wait before the attempt after attempt, or false when err must not be retried.
func (b ExponentialBackoff) Delay(attempt int, err error) (time.Duration, bool) {
	if !retryable(err) {
		return 0, false
	}
	var apiErr *tgbotapi.Error
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusTooManyRequests && apiErr.RetryAfter > 0 {
		wait := time.Duration(apiErr.RetryAfter) * time.Second
		return wait, wait <= b.Ceiling
	}
	return b.doubled(attempt), true
}

// doubled returns Base doubled attempt-1 times, stopping at Ceiling. It multiplies instead of
// shifting, and stops once the next doubling would reach Ceiling, so it cannot overflow.
func (b ExponentialBackoff) doubled(attempt int) time.Duration {
	wait := b.Base
	for n := 1; n < attempt && wait < b.Ceiling; n++ {
		if wait > b.Ceiling/2 {
			return b.Ceiling
		}
		wait *= 2
	}
	return min(wait, b.Ceiling)
}

// retryable reports whether err is worth another attempt: a transport error or an undecodable
// body (anything that is not a *tgbotapi.Error), a 429, or a 5xx. Every other answer is final.
func retryable(err error) bool {
	var apiErr *tgbotapi.Error
	if !errors.As(err, &apiErr) {
		return true
	}
	return apiErr.Code == http.StatusTooManyRequests || apiErr.Code >= http.StatusInternalServerError
}

// RetryingSender decorates a sender with the retry policy: at most maxAttempts attempts, with the
// waits chosen by backoff. Waits go through sleep, so tests can record them instead of sleeping.
type RetryingSender struct {
	next        sender
	backoff     Delayer
	maxAttempts int
	sleep       func(time.Duration)
	logger      *zap.Logger
}

// NewRetryingSender wraps next with the retry policy described on RetryingSender.
func NewRetryingSender(next sender, backoff Delayer, maxAttempts int, sleep func(time.Duration), logger *zap.Logger) RetryingSender {
	return RetryingSender{next: next, backoff: backoff, maxAttempts: maxAttempts, sleep: sleep, logger: logger}
}

// withLogger returns a copy of r that logs through logger.
func (r RetryingSender) withLogger(logger *zap.Logger) RetryingSender {
	r.logger = logger
	return r
}

// Send sends c and retries it as the backoff allows. When it gives up, the error reports the
// number of attempts made.
func (r RetryingSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	for attempt := 1; ; attempt++ {
		msg, err := r.next.Send(c)
		if err == nil {
			return msg, nil
		}
		wait, retry := r.nextWait(attempt, err)
		if !retry {
			return tgbotapi.Message{}, fmt.Errorf("failed to send message after %d attempts: %w", attempt, err)
		}
		r.sleep(wait)
	}
}

// nextWait logs the failed attempt and returns the wait before the next one, or false when no
// attempt is left or the error is not retried.
func (r RetryingSender) nextWait(attempt int, err error) (time.Duration, bool) {
	r.logger.Warn("Failed to send Telegram message",
		zap.Int("attempt", attempt),
		zap.Int("maxAttempts", r.maxAttempts),
		zap.Error(err))
	if attempt >= r.maxAttempts {
		return 0, false
	}
	wait, ok := r.backoff.Delay(attempt, err)
	if ok {
		r.logger.Info("Retrying Telegram message send", zap.Duration("backoff", wait))
	}
	return wait, ok
}
