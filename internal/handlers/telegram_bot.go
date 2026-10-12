package handlers

import (
	"context"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
)

// The interfaces below are what the command types need. They are consumer-defined: each one lists
// only the methods its caller uses.

// WorkflowDispatcher starts a GitHub Actions workflow.
type WorkflowDispatcher interface {
	DispatchWorkflow(ctx context.Context, owner, repo, workflowFile, ref string) (string, error)
}

// DeploymentRestarter restarts a Deployment and waits for its rollout.
type DeploymentRestarter interface {
	RestartDeployment(ctx context.Context, namespace, deployment string) error
	WaitForRollout(ctx context.Context, namespace, deployment string, pollInterval time.Duration) error
}

// JobCreator creates a Job from a CronJob.
type JobCreator interface {
	CreateJobFromCronJob(ctx context.Context, namespace, name string) (string, error)
}

// KubeClient is the Kubernetes client that serves both restarts and jobs.
type KubeClient interface {
	DeploymentRestarter
	JobCreator
}

// TelegramMessenger sends a text to a chat.
type TelegramMessenger interface {
	SendMessage(string, string) error
}

// BotHandlerDeps holds the clients and the command configuration of the bot. A client may be nil
// when no configured command uses it. The settings left at zero take their defaults.
type BotHandlerDeps struct {
	GitHubClient   WorkflowDispatcher
	KubeClient     DeploymentRestarter
	TelegramClient TelegramMessenger
	JobClient      JobCreator
	Workflows      []config.WorkflowCommandConfig
	Restarts       []config.RestartCommandConfig
	Jobs           []config.JobCommandConfig
	GitHubWebURL   string
	RestartReady   config.RestartReadyConfig
	Catalog        *messages.Catalog
	Logger         *zap.Logger
}

// BotHandler routes the command updates received through Telegram long polling. It checks that
// the command exists and that the chat may run it, then hands it to the command's spec.
type BotHandler struct {
	replier
	specs map[string]CommandSpec
}

func NewBotHandler(deps BotHandlerDeps) *BotHandler {
	d := deps.withDefaults()
	specs := make(map[string]CommandSpec)
	for _, spec := range NewCommandSpecs(d) {
		specs[spec.Name()] = spec
	}
	return &BotHandler{
		replier: newReplier(d),
		specs:   specs,
	}
}

func (h *BotHandler) Handle(update tgbotapi.Update) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	commandName := update.Message.Command()
	chatID := strconv.FormatInt(update.Message.Chat.ID, 10)

	spec, exists := h.specs[commandName]
	if !exists {
		h.reply(chatID, "bot.unknown_command", nil)
		return
	}
	if !spec.Authorize(chatID) {
		h.logger.Warn("Unauthorized Telegram workflow command attempt",
			zap.String("command", commandName),
			zap.String("chat_id", chatID),
			zap.String("username", username(update.Message)))
		h.reply(chatID, "bot.unauthorized", nil)
		return
	}
	spec.Run(context.Background(), chatID)
}

func username(message *tgbotapi.Message) string {
	if message.From == nil {
		return ""
	}
	return message.From.UserName
}
