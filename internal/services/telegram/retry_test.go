package telegram

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// scriptedSender returns errs[i] on call i, and succeeds once the script is exhausted.
type scriptedSender struct {
	errs  []error
	calls int
}

func (s *scriptedSender) Send(tgbotapi.Chattable) (tgbotapi.Message, error) {
	i := s.calls
	s.calls++
	if i < len(s.errs) {
		return tgbotapi.Message{}, s.errs[i]
	}
	return tgbotapi.Message{}, nil
}

// apiError is the error tgbotapi returns for a Telegram error answer with the given code.
func apiError(code int) error {
	return &tgbotapi.Error{Code: code, Message: "answer"}
}

// rateLimited is a 429 answer carrying retry_after seconds.
func rateLimited(retryAfter int) error {
	return &tgbotapi.Error{
		Code:               429,
		Message:            "Too Many Requests",
		ResponseParameters: tgbotapi.ResponseParameters{RetryAfter: retryAfter},
	}
}

// retryTest runs a RetryingSender over src and returns the sender's call count, the recorded
// waits and the error of Send.
func retryTest(src *scriptedSender, maxAttempts int, backoff Delayer) (int, []time.Duration, error) {
	rec := &sleepRecorder{}
	r := NewRetryingSender(src, backoff, maxAttempts, rec.sleep, zap.NewNop())
	_, err := r.Send(tgbotapi.NewMessage(1, "x"))
	return src.calls, rec.waits, err
}

func testBackoff() ExponentialBackoff {
	return ExponentialBackoff{Base: time.Second, Ceiling: 30 * time.Second}
}

func TestRetryingSenderClassification(t *testing.T) {
	tests := []struct {
		name        string
		errs        []error
		maxAttempts int
		wantCalls   int
		wantWaits   []time.Duration // checked only when set
		wantErr     string          // substring of the error; empty means success
	}{
		{name: "400 is not retried", errs: []error{apiError(400), nil}, maxAttempts: 3, wantCalls: 1, wantErr: "failed to send message after 1 attempts"},
		{name: "429 is retried", errs: []error{apiError(429), nil}, maxAttempts: 3, wantCalls: 2},
		{name: "500 is retried", errs: []error{apiError(500), nil}, maxAttempts: 3, wantCalls: 2},
		{name: "502 is retried", errs: []error{apiError(502), nil}, maxAttempts: 3, wantCalls: 2},
		{name: "network error is retried", errs: []error{errors.New("connection reset"), nil}, maxAttempts: 3, wantCalls: 2, wantWaits: []time.Duration{time.Second}},
		{name: "max_attempts 1 disables retries", errs: []error{apiError(502), nil}, maxAttempts: 1, wantCalls: 1, wantErr: "failed to send message after 1 attempts"},
		{name: "attempts exhausted", errs: []error{apiError(502), apiError(502), apiError(502), nil}, maxAttempts: 3, wantCalls: 3, wantErr: "failed to send message after 3 attempts"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls, waits, err := retryTest(&scriptedSender{errs: tt.errs}, tt.maxAttempts, testBackoff())
			if calls != tt.wantCalls {
				t.Errorf("attempts = %d, want %d", calls, tt.wantCalls)
			}
			if tt.wantWaits != nil && !slices.Equal(waits, tt.wantWaits) {
				t.Errorf("waits = %v, want %v", waits, tt.wantWaits)
			}
			switch {
			case tt.wantErr == "" && err != nil:
				t.Errorf("error = %v, want nil", err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Errorf("error = %v, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRetryingSenderBackoffSequence(t *testing.T) {
	calls, waits, err := retryTest(&scriptedSender{errs: []error{apiError(502), apiError(502), nil}}, 3, testBackoff())
	if err != nil {
		t.Fatalf("error = %v, want success on the third attempt", err)
	}
	if calls != 3 {
		t.Errorf("attempts = %d, want 3", calls)
	}
	if want := []time.Duration{time.Second, 2 * time.Second}; !slices.Equal(waits, want) {
		t.Errorf("waits = %v, want %v", waits, want)
	}
}

func TestBackoffNoOverflowAtHighAttempts(t *testing.T) {
	backoff := testBackoff()
	for _, attempt := range []int{1, 35, 64, 100, 1000} {
		wait, ok := backoff.Delay(attempt, apiError(502))
		if !ok || wait <= 0 || wait > backoff.Ceiling {
			t.Errorf("attempt %d: Delay() = %s, %v; want a wait in (0, %s]", attempt, wait, ok, backoff.Ceiling)
		}
		if attempt >= 64 && wait != backoff.Ceiling {
			t.Errorf("attempt %d: wait = %s, want the ceiling %s", attempt, wait, backoff.Ceiling)
		}
	}
}

func TestRetryingSenderCapsBackoffAtCeiling(t *testing.T) {
	backoff := ExponentialBackoff{Base: time.Second, Ceiling: 1500 * time.Millisecond}
	_, waits, _ := retryTest(&scriptedSender{errs: []error{apiError(502), apiError(502), apiError(502), nil}}, 4, backoff)
	if want := []time.Duration{time.Second, 1500 * time.Millisecond, 1500 * time.Millisecond}; !slices.Equal(waits, want) {
		t.Errorf("waits = %v, want %v", waits, want)
	}
}

func TestRetryingSenderHonoursRetryAfter(t *testing.T) {
	calls, waits, err := retryTest(&scriptedSender{errs: []error{rateLimited(5), nil}}, 3, testBackoff())
	if err != nil {
		t.Fatalf("error = %v, want success", err)
	}
	if calls != 2 {
		t.Errorf("attempts = %d, want 2", calls)
	}
	if want := []time.Duration{5 * time.Second}; !slices.Equal(waits, want) {
		t.Errorf("waits = %v, want %v", waits, want)
	}
}

func TestRetryingSenderStopsAboveMaxRetryAfter(t *testing.T) {
	calls, waits, err := retryTest(&scriptedSender{errs: []error{rateLimited(120), nil}}, 3, testBackoff())
	if err == nil {
		t.Fatal("error = nil, want the rate limit error")
	}
	if calls != 1 {
		t.Errorf("attempts = %d, want 1", calls)
	}
	if len(waits) != 0 {
		t.Errorf("waits = %v, want none", waits)
	}
}
