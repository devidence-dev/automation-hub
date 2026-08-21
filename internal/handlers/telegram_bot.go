package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"automation-hub/internal/config"
)

// runLookupAttempts/runLookupDelay bound how long dispatchWorkflow waits for
// the dispatched run to show up in GitHub's run list before falling back to
// the repository's general Actions URL.
const (
	runLookupAttempts = 5
	runLookupDelay    = 2 * time.Second
)

// restartReadyTimeout/restartReadyPollInterval bound how long restartDeployment
// waits, in the background, for the rollout to report ready before giving up
// and telling the user to check it manually.
const (
	restartReadyTimeout      = 3 * time.Minute
	restartReadyPollInterval = 5 * time.Second
)

type workflowDispatcher interface {
	DispatchWorkflow(ctx context.Context, owner, repo, workflowFile, ref string) error
	FindLatestRunURL(ctx context.Context, owner, repo, workflowFile string, after time.Time) (string, error)
}

type deploymentRestarter interface {
	RestartDeployment(ctx context.Context, namespace, deployment string) error
	WaitForRollout(ctx context.Context, namespace, deployment string, pollInterval time.Duration) error
}

type telegramMessenger interface {
	SendMessage(string, string) error
}

// BotHandler processes command updates received through Telegram long polling.
type BotHandler struct {
	githubClient   workflowDispatcher
	kubeClient     deploymentRestarter
	telegramClient telegramMessenger
	commands       map[string]config.WorkflowCommandConfig
	restarts       map[string]config.RestartCommandConfig
	logger         *zap.Logger
}

func NewBotHandler(githubClient workflowDispatcher, kubeClient deploymentRestarter, telegramClient telegramMessenger, workflows []config.WorkflowCommandConfig, restarts []config.RestartCommandConfig, logger *zap.Logger) *BotHandler {
	commands := make(map[string]config.WorkflowCommandConfig, len(workflows))
	for _, workflow := range workflows {
		commands[workflow.Command] = workflow
	}
	restartCommands := make(map[string]config.RestartCommandConfig, len(restarts))
	for _, restart := range restarts {
		restartCommands[restart.Command] = restart
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BotHandler{
		githubClient:   githubClient,
		kubeClient:     kubeClient,
		telegramClient: telegramClient,
		commands:       commands,
		restarts:       restartCommands,
		logger:         logger,
	}
}

func (h *BotHandler) Handle(update tgbotapi.Update) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	commandName := update.Message.Command()
	chatID := strconv.FormatInt(update.Message.Chat.ID, 10)

	if workflow, exists := h.commands[commandName]; exists {
		if !isAllowed(chatID, workflow.AllowedChatIDs) {
			h.handleUnauthorizedCommand(commandName, chatID, username(update.Message))
			return
		}
		h.dispatchWorkflow(commandName, chatID, workflow)
		return
	}

	if restart, exists := h.restarts[commandName]; exists {
		if !isAllowed(chatID, restart.AllowedChatIDs) {
			h.handleUnauthorizedCommand(commandName, chatID, username(update.Message))
			return
		}
		h.restartDeployment(commandName, chatID, restart)
		return
	}

	h.sendMessage(chatID, "❓ Unknown command.")
}

func (h *BotHandler) handleUnauthorizedCommand(commandName, chatID, username string) {
	h.logger.Warn("Unauthorized Telegram workflow command attempt",
		zap.String("command", commandName),
		zap.String("chat_id", chatID),
		zap.String("username", username))
	h.sendMessage(chatID, "🚫 You are not authorized to run this command.")
}

func (h *BotHandler) dispatchWorkflow(commandName, chatID string, workflow config.WorkflowCommandConfig) {
	dispatchedAt := time.Now().UTC()
	if err := h.githubClient.DispatchWorkflow(context.Background(), workflow.Owner, workflow.Repo, workflow.WorkflowFile, workflow.Ref); err != nil {
		h.logger.Error("Failed to dispatch GitHub workflow", zap.Error(err), zap.String("command", commandName), zap.String("chat_id", chatID))
		h.sendMessage(chatID, "❌ Failed to dispatch the workflow. Check the hub logs or try again later.")
		return
	}

	runURL := h.findRunURL(commandName, workflow, dispatchedAt)
	h.sendMessage(chatID, fmt.Sprintf("🚀 Workflow *%s* dispatched successfully. View it at %s", workflow.WorkflowFile, runURL))
}

func (h *BotHandler) restartDeployment(commandName, chatID string, restart config.RestartCommandConfig) {
	if err := h.kubeClient.RestartDeployment(context.Background(), restart.Namespace, restart.Deployment); err != nil {
		h.logger.Error("Failed to restart deployment", zap.Error(err), zap.String("command", commandName), zap.String("chat_id", chatID))
		h.sendMessage(chatID, "❌ Failed to restart the deployment. Check the hub logs or try again later.")
		return
	}
	h.sendMessage(chatID, fmt.Sprintf("🔄 Deployment *%s* restart triggered. Waiting for it to come back online...", restart.Deployment))
	go h.awaitRestartReady(commandName, chatID, restart)
}

// awaitRestartReady runs in the background so it never blocks the polling
// loop from handling the next Telegram update while it waits for the new
// pod to become ready.
func (h *BotHandler) awaitRestartReady(commandName, chatID string, restart config.RestartCommandConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), restartReadyTimeout)
	defer cancel()

	if err := h.kubeClient.WaitForRollout(ctx, restart.Namespace, restart.Deployment, restartReadyPollInterval); err != nil {
		h.logger.Warn("Deployment did not report ready in time", zap.Error(err), zap.String("command", commandName), zap.String("chat_id", chatID))
		h.sendMessage(chatID, fmt.Sprintf("⏱️ Deployment *%s* did not report ready within %s — check it manually.", restart.Deployment, restartReadyTimeout))
		return
	}
	h.sendMessage(chatID, fmt.Sprintf("✅ Deployment *%s* is up and running again.", restart.Deployment))
}

// findRunURL polls GitHub for the run that the dispatch just created.
// workflow_dispatch itself never returns the run's identity, so this waits a
// few seconds for the run to appear before giving up and pointing at the
// repository's general Actions tab instead.
func (h *BotHandler) findRunURL(commandName string, workflow config.WorkflowCommandConfig, dispatchedAt time.Time) string {
	actionsURL := fmt.Sprintf("https://github.com/%s/%s/actions", workflow.Owner, workflow.Repo)
	for attempt := 0; attempt < runLookupAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(runLookupDelay)
		}
		runURL, err := h.githubClient.FindLatestRunURL(context.Background(), workflow.Owner, workflow.Repo, workflow.WorkflowFile, dispatchedAt)
		if err != nil {
			h.logger.Warn("Failed to look up dispatched workflow run", zap.Error(err), zap.String("command", commandName))
			return actionsURL
		}
		if runURL != "" {
			return runURL
		}
	}
	return actionsURL
}

func username(message *tgbotapi.Message) string {
	if message.From == nil {
		return ""
	}
	return message.From.UserName
}

func (h *BotHandler) sendMessage(chatID, message string) {
	if h.telegramClient == nil {
		return
	}
	if err := h.telegramClient.SendMessage(chatID, message); err != nil {
		h.logger.Error("Failed to reply to Telegram command", zap.String("chat_id", chatID), zap.Error(err))
	}
}

func isAllowed(chatID string, allowedChatIDs []string) bool {
	for _, allowedChatID := range allowedChatIDs {
		if allowedChatID == chatID {
			return true
		}
	}
	return false
}
