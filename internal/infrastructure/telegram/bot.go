package telegram

import (
	"context"
	"fmt"
	"log"

	"dutch-learning-bot/internal/interfaces/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot wraps the Telegram bot API
type Bot struct {
	api        *tgbotapi.BotAPI
	dispatcher *defaultDispatcher
}

// NewBot creates a new bot instance
func NewBot(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &Bot{
		api:        api,
		dispatcher: newDefaultDispatcher(),
	}, nil
}

// GetAPI returns the underlying bot API
func (b *Bot) GetAPI() *tgbotapi.BotAPI {
	return b.api
}

// GetDispatcher returns the bot's dispatcher
func (b *Bot) GetDispatcher() telegram.Dispatcher {
	return b.dispatcher
}

// GetUpdatesChan returns a channel for receiving updates
func (b *Bot) GetUpdatesChan() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	return updates
}

// SendMessage sends a text message
func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

// SendMessageWithMarkdown sends a message with markdown formatting
func (b *Bot) SendMessageWithMarkdown(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.api.Send(msg)
	return err
}

// SendMessageWithKeyboard sends a message with inline keyboard
func (b *Bot) SendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard
	_, err := b.api.Send(msg)
	return err
}

// EditMessage edits a message
func (b *Bot) EditMessage(chatID int64, messageID int, text string) error {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to edit message: %w", err)
	}
	return nil
}

// EditMessageWithKeyboard edits an existing message and adds a keyboard
func (b *Bot) EditMessageWithKeyboard(chatID int64, messageID int, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard
	_, err := b.api.Send(edit)
	if err != nil {
		log.Printf("Failed to edit message with keyboard: %v", err)
		return fmt.Errorf("failed to edit message with keyboard: %w", err)
	}
	return err
}

// AnswerCallbackQuery answers a callback query
func (b *Bot) AnswerCallbackQuery(callbackID string, text string) error {
	callback := tgbotapi.NewCallback(callbackID, text)
	_, err := b.api.Request(callback)
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}
	return nil
}

// SetupCommands sets up bot commands
func (b *Bot) SetupCommands() error {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "menu", Description: "Show main menu"},
		{Command: "learn", Description: "Start learning session"},
		{Command: "stats", Description: "Show your learning statistics"},
		{Command: "settings", Description: "Show settings"},
		{Command: "help", Description: "Show help"},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := b.api.Request(config)
	if err != nil {
		return fmt.Errorf("failed to set commands: %w", err)
	}

	return nil
}

// defaultDispatcher implements the Dispatcher interface
type defaultDispatcher struct {
	handlers map[string]telegram.HandlerFunc
}

// newDefaultDispatcher creates a new default dispatcher
func newDefaultDispatcher() *defaultDispatcher {
	return &defaultDispatcher{
		handlers: make(map[string]telegram.HandlerFunc),
	}
}

// RegisterHandler registers a handler for a specific command
func (d *defaultDispatcher) RegisterHandler(command string, handler telegram.HandlerFunc) {
	d.handlers[command] = handler
}

// Dispatch dispatches an update to the appropriate handler
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
