package handlers

import (
	"context"

	"go.uber.org/zap"
)

// ScheduledRunner runs scheduled commands outside any chat. It logs each outcome and, when a run
// fails, tells every chat of the command's allow list.
type ScheduledRunner struct {
	replier
}

func NewScheduledRunner(deps BotHandlerDeps) *ScheduledRunner {
	return &ScheduledRunner{replier: newReplier(deps.withDefaults())}
}

// Wrap returns the function the scheduler calls for spec.
func (s *ScheduledRunner) Wrap(spec CommandSpec) func() {
	return func() { s.run(spec) }
}

func (s *ScheduledRunner) run(spec CommandSpec) {
	fields, err := spec.RunScheduled(context.Background())
	if err != nil {
		s.logger.Error("Scheduled command failed", append(fields, zap.Error(err))...)
		for _, chatID := range spec.Chats() {
			s.reply(chatID, spec.FailureKey(), map[string]string{"command": spec.Name()})
		}
		return
	}
	s.logger.Info("Scheduled command finished", fields...)
}
