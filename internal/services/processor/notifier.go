package processor

import "fmt"

// messageSender is the outbound side a processor needs: it delivers text to a chat.
// *telegram.Client satisfies it.
type messageSender interface {
	SendMessage(chatID, message string) error
}

// TemplateNotifier fills a template with its arguments and sends the result to one chat.
type TemplateNotifier struct {
	sender   messageSender
	chatID   string
	template string
}

// Notify formats the template with args and sends it.
func (n TemplateNotifier) Notify(args ...any) error {
	return n.sender.SendMessage(n.chatID, fmt.Sprintf(n.template, args...))
}
