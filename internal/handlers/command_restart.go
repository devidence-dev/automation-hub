package handlers

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"automation-hub/internal/config"
)

// errRestartNotScheduled is returned for a restart asked to run from the scheduler. Restarts have no schedule.
var errRestartNotScheduled = errors.New("restart commands have no schedule")

// restartCommand restarts a Deployment and then waits, in the background, for it to report ready.
type restartCommand struct {
	command
	restart   config.RestartCommandConfig
	restarter DeploymentRestarter
	ready     config.RestartReadyConfig
}

func newRestartCommand(deps BotHandlerDeps, restart config.RestartCommandConfig) *restartCommand {
	return &restartCommand{
		command:   newCommand(deps, restart.Command, restart.Description, "", "", restart.AllowedChatIDs),
		restart:   restart,
		restarter: deps.KubeClient,
		ready:     deps.RestartReady,
	}
}

// Run triggers the rollout restart. The readiness wait runs in the background, so the bot
// handles the next update while the new pods start.
func (r *restartCommand) Run(ctx context.Context, chatID string) {
	if err := r.restarter.RestartDeployment(ctx, r.restart.Namespace, r.restart.Deployment); err != nil {
		r.logger.Error("Failed to restart deployment", zap.Error(err), zap.String("command", r.name), zap.String("chat_id", chatID))
		r.reply(chatID, "bot.restart_failed", nil)
		return
	}
	r.reply(chatID, "bot.restart_triggered", map[string]string{"deployment": r.restart.Deployment})
	go r.awaitReady(ctx, chatID)
}

// RunScheduled always fails: a restart has no schedule.
func (r *restartCommand) RunScheduled(context.Context) ([]zap.Field, error) {
	return []zap.Field{zap.String("command", r.name)}, errRestartNotScheduled
}

// awaitReady waits for the rollout up to the configured timeout. It keeps the update context
// values but not its cancellation, so the wait outlives the update that started it.
func (r *restartCommand) awaitReady(ctx context.Context, chatID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.ready.Timeout)
	defer cancel()

	if err := r.restarter.WaitForRollout(ctx, r.restart.Namespace, r.restart.Deployment, r.ready.PollInterval); err != nil {
		r.logger.Warn("Deployment did not report ready in time", zap.Error(err), zap.String("command", r.name), zap.String("chat_id", chatID))
		r.reply(chatID, "bot.restart_timeout", map[string]string{"deployment": r.restart.Deployment, "timeout": r.ready.Timeout.String()})
		return
	}
	r.reply(chatID, "bot.restart_ready", map[string]string{"deployment": r.restart.Deployment})
}
