package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandlerFunc is a function that handles a Telegram update
type HandlerFunc func(ctx context.Context, update tgbotapi.Update) error

// Dispatcher handles routing of Telegram updates to appropriate handlers
type Dispatcher interface {
	// RegisterHandler registers a handler for a specific command
	RegisterHandler(command string, handler HandlerFunc)
	// Dispatch dispatches an update to the appropriate handler
	Dispatch(ctx context.Context, update tgbotapi.Update) error
}

// NewDispatcher creates a new dispatcher instance
func NewDispatcher() Dispatcher {
	return &defaultDispatcher{
		handlers: make(map[string]HandlerFunc),
	}
}

type defaultDispatcher struct {
	handlers map[string]HandlerFunc
}

func (d *defaultDispatcher) RegisterHandler(command string, handler HandlerFunc) {
	d.handlers[command] = handler
}

func (d *defaultDispatcher) Dispatch(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	command := update.Message.Command()
	if command == "" {
		return nil
	}

	handler, exists := d.handlers[command]
	if !exists {
		return nil
	}

	return handler(ctx, update)
}
