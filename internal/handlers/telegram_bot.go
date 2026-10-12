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

// restartReadyTimeout/restartReadyPollInterval bound how long restartDeployment
// waits, in the background, for the rollout to report ready before giving up
// and telling the user to check it manually.
const (
	restartReadyTimeout      = 3 * time.Minute
	restartReadyPollInterval = 5 * time.Second
)

type workflowDispatcher interface {
	DispatchWorkflow(ctx context.Context, owner, repo, workflowFile, ref string) (string, error)
}

type deploymentRestarter interface {
	RestartDeployment(ctx context.Context, namespace, deployment string) error
	WaitForRollout(ctx context.Context, namespace, deployment string, pollInterval time.Duration) error
}

type jobCreator interface {
	CreateJobFromCronJob(ctx context.Context, namespace, name string) (string, error)
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
	jobClient      jobCreator
	jobs           map[string]config.JobCommandConfig
	logger         *zap.Logger
}

// BotHandlerDeps holds the clients and command configuration of a BotHandler.
// A client may be nil when no configured command uses it.
type BotHandlerDeps struct {
	GitHubClient   workflowDispatcher
	KubeClient     deploymentRestarter
	TelegramClient telegramMessenger
	JobClient      jobCreator
	Workflows      []config.WorkflowCommandConfig
	Restarts       []config.RestartCommandConfig
	Jobs           []config.JobCommandConfig
	Logger         *zap.Logger
}

func NewBotHandler(deps BotHandlerDeps) *BotHandler {
	commands := make(map[string]config.WorkflowCommandConfig, len(deps.Workflows))
	for _, workflow := range deps.Workflows {
		commands[workflow.Command] = workflow
	}
	restartCommands := make(map[string]config.RestartCommandConfig, len(deps.Restarts))
	for _, restart := range deps.Restarts {
		restartCommands[restart.Command] = restart
	}
	jobCommands := make(map[string]config.JobCommandConfig, len(deps.Jobs))
	for _, job := range deps.Jobs {
		jobCommands[job.Command] = job
	}
	logger := deps.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &BotHandler{
		githubClient:   deps.GitHubClient,
		kubeClient:     deps.KubeClient,
		telegramClient: deps.TelegramClient,
		commands:       commands,
		restarts:       restartCommands,
		jobClient:      deps.JobClient,
		jobs:           jobCommands,
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

	if job, exists := h.jobs[commandName]; exists {
		if !isAllowed(chatID, job.AllowedChatIDs) {
			h.handleUnauthorizedCommand(commandName, chatID, username(update.Message))
			return
		}
		h.createJob(commandName, chatID, job)
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
	runURL, err := h.RunWorkflow(context.Background(), workflow)
	if err != nil {
		h.logger.Error("Failed to dispatch GitHub workflow", zap.Error(err), zap.String("command", commandName), zap.String("chat_id", chatID))
		h.sendMessage(chatID, "❌ Failed to dispatch the workflow. Check the hub logs or try again later.")
		return
	}
	h.sendMessage(chatID, fmt.Sprintf("🚀 Workflow *%s* dispatched successfully. View it at %s", workflow.WorkflowFile, runURL))
}

// RunWorkflow dispatches the workflow and returns the URL of the run GitHub
// created. If GitHub returns no run details, it returns the repository's
// Actions page instead. It does not depend on a chat, so Telegram commands and
// schedules share it.
func (h *BotHandler) RunWorkflow(ctx context.Context, workflow config.WorkflowCommandConfig) (string, error) {
	runURL, err := h.githubClient.DispatchWorkflow(ctx, workflow.Owner, workflow.Repo, workflow.WorkflowFile, workflow.Ref)
	if err != nil {
		return "", err
	}
	if runURL == "" {
		runURL = fmt.Sprintf("https://github.com/%s/%s/actions", workflow.Owner, workflow.Repo)
	}
	return runURL, nil
}

// RunScheduled runs a workflow from the scheduler, outside any Telegram update.
// It logs the outcome. On failure it tells every chat in allowed_chat_ids.
func (h *BotHandler) RunScheduled(workflow config.WorkflowCommandConfig) {
	fields := []zap.Field{
		zap.String("command", workflow.Command),
		zap.String("owner", workflow.Owner),
		zap.String("repo", workflow.Repo),
		zap.String("workflow", workflow.WorkflowFile),
	}
	runURL, err := h.RunWorkflow(context.Background(), workflow)
	if err != nil {
		h.logger.Error("Scheduled workflow dispatch failed", append(fields, zap.String("result", "failed"), zap.Error(err))...)
		for _, chatID := range workflow.AllowedChatIDs {
			h.sendMessage(chatID, fmt.Sprintf("❌ Scheduled run of `%s` failed to dispatch. Check the hub logs or try again later.", workflow.Command))
		}
		return
	}
	h.logger.Info("Scheduled workflow dispatched", append(fields, zap.String("result", "dispatched"), zap.String("run_url", runURL))...)
}

func (h *BotHandler) createJob(commandName, chatID string, job config.JobCommandConfig) {
	jobName, err := h.RunJob(context.Background(), job)
	if err != nil {
		h.logger.Error("Failed to create Kubernetes job", zap.Error(err), zap.String("command", commandName), zap.String("chat_id", chatID))
		h.sendMessage(chatID, "❌ Failed to create the job. Check the hub logs or try again later.")
		return
	}
	h.sendMessage(chatID, fmt.Sprintf("🚀 Job *%s* created from *%s*.", jobName, job.CronJob))
}

// RunJob creates a Job from the entry's CronJob and returns its name. Like
// RunWorkflow it does not depend on a chat, so commands and schedules share it.
func (h *BotHandler) RunJob(ctx context.Context, job config.JobCommandConfig) (string, error) {
	return h.jobClient.CreateJobFromCronJob(ctx, job.Namespace, job.CronJob)
}

// RunScheduledJob runs a job command from the scheduler. It logs the outcome.
// On failure it tells every chat in allowed_chat_ids.
func (h *BotHandler) RunScheduledJob(job config.JobCommandConfig) {
	fields := []zap.Field{
		zap.String("command", job.Command),
		zap.String("namespace", job.Namespace),
		zap.String("cronjob", job.CronJob),
	}
	jobName, err := h.RunJob(context.Background(), job)
	if err != nil {
		h.logger.Error("Scheduled job creation failed", append(fields, zap.String("result", "failed"), zap.Error(err))...)
		for _, chatID := range job.AllowedChatIDs {
			h.sendMessage(chatID, fmt.Sprintf("❌ Scheduled run of `%s` failed to create the job. Check the hub logs or try again later.", job.Command))
		}
		return
	}
	h.logger.Info("Scheduled job created", append(fields, zap.String("result", "created"), zap.String("job", jobName))...)
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
