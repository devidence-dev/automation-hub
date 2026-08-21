package handlers

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"automation-hub/internal/config"
)

type workflowDispatcher interface {
	DispatchWorkflow(context.Context, string, string, string, string) error
}

type telegramMessenger interface {
	SendMessage(string, string) error
}

// BotHandler processes command updates received through Telegram long polling.
type BotHandler struct {
	githubClient   workflowDispatcher
	telegramClient telegramMessenger
	commands       map[string]config.WorkflowCommandConfig
	logger         *zap.Logger
}

func NewBotHandler(githubClient workflowDispatcher, telegramClient telegramMessenger, workflows []config.WorkflowCommandConfig, logger *zap.Logger) *BotHandler {
	commands := make(map[string]config.WorkflowCommandConfig, len(workflows))
	for _, workflow := range workflows {
		commands[workflow.Command] = workflow
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BotHandler{githubClient: githubClient, telegramClient: telegramClient, commands: commands, logger: logger}
}

func (h *BotHandler) Handle(update tgbotapi.Update) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	commandName := update.Message.Command()
	workflow, exists := h.commands[commandName]
	chatID := strconv.FormatInt(update.Message.Chat.ID, 10)
	if !exists {
		h.sendMessage(chatID, "Comando no reconocido.")
		return
	}
	if !isAllowed(chatID, workflow.AllowedChatIDs) {
		username := ""
		if update.Message.From != nil {
			username = update.Message.From.UserName
		}
		h.logger.Warn("Unauthorized Telegram workflow command attempt",
			zap.String("command", commandName),
			zap.String("chat_id", chatID),
			zap.String("username", username))
		h.sendMessage(chatID, "No estás autorizado para ejecutar este comando.")
		return
	}

	if err := h.githubClient.DispatchWorkflow(context.Background(), workflow.Owner, workflow.Repo, workflow.WorkflowFile, workflow.Ref); err != nil {
		h.logger.Error("Failed to dispatch GitHub workflow", zap.Error(err), zap.String("command", commandName), zap.String("chat_id", chatID))
		h.sendMessage(chatID, "No se pudo disparar el workflow. Revisa los logs del hub o inténtalo de nuevo más tarde.")
		return
	}

	actionsURL := fmt.Sprintf("https://github.com/%s/%s/actions", workflow.Owner, workflow.Repo)
	h.sendMessage(chatID, fmt.Sprintf("Workflow *%s* disparado correctamente. Puedes verlo en %s", workflow.WorkflowFile, actionsURL))
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
