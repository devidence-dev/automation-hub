package handlers

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"automation-hub/internal/config"
)

// workflowCommand dispatches a GitHub Actions workflow.
type workflowCommand struct {
	command
	workflow   config.WorkflowCommandConfig
	dispatcher WorkflowDispatcher
	webURL     string
}

func newWorkflowCommand(deps BotHandlerDeps, workflow config.WorkflowCommandConfig) *workflowCommand {
	return &workflowCommand{
		command:    newCommand(deps, workflow.Command, workflow.Description, workflow.Schedule, "bot.scheduled_dispatch_failed", workflow.AllowedChatIDs),
		workflow:   workflow,
		dispatcher: deps.GitHubClient,
		webURL:     deps.GitHubWebURL,
	}
}

// Run dispatches the workflow and replies with its run link.
func (w *workflowCommand) Run(ctx context.Context, chatID string) {
	runURL, err := w.dispatch(ctx)
	if err != nil {
		w.logger.Error("Failed to dispatch GitHub workflow", zap.Error(err), zap.String("command", w.name), zap.String("chat_id", chatID))
		w.reply(chatID, "bot.dispatch_failed", nil)
		return
	}
	w.reply(chatID, "bot.workflow_dispatched", map[string]string{"workflow": w.workflow.WorkflowFile, "run_url": runURL})
}

// RunScheduled dispatches the workflow from the scheduler.
func (w *workflowCommand) RunScheduled(ctx context.Context) ([]zap.Field, error) {
	fields := []zap.Field{
		zap.String("command", w.name),
		zap.String("owner", w.workflow.Owner),
		zap.String("repo", w.workflow.Repo),
		zap.String("workflow", w.workflow.WorkflowFile),
	}
	runURL, err := w.dispatch(ctx)
	if err != nil {
		return append(fields, zap.String("result", "failed")), err
	}
	return append(fields, zap.String("result", "dispatched"), zap.String("run_url", runURL)), nil
}

// dispatch returns the URL of the run GitHub created. If GitHub returns no run details, it
// returns the repository's Actions page under the configured web URL.
func (w *workflowCommand) dispatch(ctx context.Context) (string, error) {
	runURL, err := w.dispatcher.DispatchWorkflow(ctx, w.workflow.Owner, w.workflow.Repo, w.workflow.WorkflowFile, w.workflow.Ref)
	if err != nil {
		return "", err
	}
	if runURL == "" {
		runURL = fmt.Sprintf("%s/%s/%s/actions", w.webURL, w.workflow.Owner, w.workflow.Repo)
	}
	return runURL, nil
}
