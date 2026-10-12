package handlers

import (
	"context"

	"go.uber.org/zap"

	"automation-hub/internal/config"
)

// jobCommand creates a Job from a CronJob template.
type jobCommand struct {
	command
	job     config.JobCommandConfig
	creator JobCreator
}

func newJobCommand(deps BotHandlerDeps, job config.JobCommandConfig) *jobCommand {
	return &jobCommand{
		command: newCommand(deps, job.Command, job.Description, job.Schedule, "bot.scheduled_job_failed", job.AllowedChatIDs),
		job:     job,
		creator: deps.JobClient,
	}
}

// Run creates the Job and replies with its name.
func (j *jobCommand) Run(ctx context.Context, chatID string) {
	jobName, err := j.create(ctx)
	if err != nil {
		j.logger.Error("Failed to create Kubernetes job", zap.Error(err), zap.String("command", j.name), zap.String("chat_id", chatID))
		j.reply(chatID, "bot.job_failed", nil)
		return
	}
	j.reply(chatID, "bot.job_created", map[string]string{"job": jobName, "cronjob": j.job.CronJob})
}

// RunScheduled creates the Job from the scheduler.
func (j *jobCommand) RunScheduled(ctx context.Context) ([]zap.Field, error) {
	fields := []zap.Field{
		zap.String("command", j.name),
		zap.String("namespace", j.job.Namespace),
		zap.String("cronjob", j.job.CronJob),
	}
	jobName, err := j.create(ctx)
	if err != nil {
		return append(fields, zap.String("result", "failed")), err
	}
	return append(fields, zap.String("result", "created"), zap.String("job", jobName)), nil
}

// create returns the name of the Job it created.
func (j *jobCommand) create(ctx context.Context) (string, error) {
	return j.creator.CreateJobFromCronJob(ctx, j.job.Namespace, j.job.CronJob)
}
