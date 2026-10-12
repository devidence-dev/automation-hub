package handlers

import (
	"context"
	"slices"

	"go.uber.org/zap"

	"automation-hub/internal/config"
	"automation-hub/internal/messages"
)

// CommandSpec is one configured bot command: a workflow, a restart or a job. The router, the
// Telegram menu and the scheduler work with specs only, so each command type is defined once.
type CommandSpec interface {
	Name() string
	Description() string
	// Schedule is the cron expression, empty when the command is not scheduled.
	Schedule() string
	// Authorize reports whether the chat may run the command.
	Authorize(chatID string) bool
	// Chats lists the chats told about a failed scheduled run.
	Chats() []string
	// FailureKey is the catalog key of that message.
	FailureKey() string
	// Run executes the command for a chat and replies in it.
	Run(ctx context.Context, chatID string)
	// RunScheduled executes the command without a chat. It returns the fields to log and the error.
	RunScheduled(ctx context.Context) ([]zap.Field, error)
}

// allowList is the allowed_chat_ids of an entry. An empty list allows no chat.
type allowList []string

func (a allowList) Authorize(chatID string) bool { return slices.Contains(a, chatID) }

func (a allowList) Chats() []string { return a }

// command holds what every command type shares.
type command struct {
	replier
	name        string
	description string
	schedule    string
	failureKey  string
	allowList
}

func newCommand(deps BotHandlerDeps, name, description, schedule, failureKey string, chats []string) command {
	return command{
		replier:     newReplier(deps),
		name:        name,
		description: description,
		schedule:    schedule,
		failureKey:  failureKey,
		allowList:   chats,
	}
}

func (c command) Name() string        { return c.name }
func (c command) Description() string { return c.description }
func (c command) Schedule() string    { return c.schedule }
func (c command) FailureKey() string  { return c.failureKey }

// replier renders a catalog text and sends it to a chat.
type replier struct {
	messenger TelegramMessenger
	catalog   *messages.Catalog
	logger    *zap.Logger
}

func newReplier(deps BotHandlerDeps) replier {
	return replier{messenger: deps.TelegramClient, catalog: deps.Catalog, logger: deps.Logger}
}

func (r replier) reply(chatID, key string, vars map[string]string) {
	text, err := r.catalog.Render(key, vars)
	if err != nil {
		r.logger.Error("Failed to render bot text", zap.String("key", key), zap.Error(err))
		return
	}
	r.send(chatID, text)
}

func (r replier) send(chatID, text string) {
	if r.messenger == nil {
		return
	}
	if err := r.messenger.SendMessage(chatID, text); err != nil {
		r.logger.Error("Failed to reply to Telegram command", zap.String("chat_id", chatID), zap.Error(err))
	}
}

// NewCommandSpecs returns one spec per configured command, in config order: workflows, restarts,
// then jobs. A client the config does not use stays a nil interface.
func NewCommandSpecs(deps BotHandlerDeps) []CommandSpec {
	d := deps.withDefaults()
	specs := make([]CommandSpec, 0, len(d.Workflows)+len(d.Restarts)+len(d.Jobs))
	for _, workflow := range d.Workflows {
		specs = append(specs, newWorkflowCommand(d, workflow))
	}
	for _, restart := range d.Restarts {
		specs = append(specs, newRestartCommand(d, restart))
	}
	for _, job := range d.Jobs {
		specs = append(specs, newJobCommand(d, job))
	}
	return specs
}

// withDefaults fills the settings a caller leaves out. The defaults are the config package's, so
// they match what a config without these keys gives.
func (d BotHandlerDeps) withDefaults() BotHandlerDeps {
	if d.Logger == nil {
		d.Logger = zap.NewNop()
	}
	if d.Catalog == nil {
		d.Catalog = defaultCatalog()
	}
	if d.GitHubWebURL == "" {
		d.GitHubWebURL = config.DefaultGitHubWebURL
	}
	if d.RestartReady.Timeout <= 0 {
		d.RestartReady.Timeout = config.DefaultRestartTimeout
	}
	if d.RestartReady.PollInterval <= 0 {
		d.RestartReady.PollInterval = config.DefaultRestartPollInterval
	}
	return d
}

// defaultCatalog returns the embedded catalog. Its load is covered by the messages tests, so a
// failure here means the embedded file is broken and the process cannot run.
func defaultCatalog() *messages.Catalog {
	cat, err := messages.Load(nil)
	if err != nil {
		panic("embedded message catalog: " + err.Error())
	}
	return cat
}
