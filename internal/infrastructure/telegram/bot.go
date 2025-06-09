package telegram

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot wraps the Telegram bot API
type Bot struct {
	api *tgbotapi.BotAPI
}

// NewBot creates a new Telegram bot
func NewBot(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = false
	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{api: api}, nil
}

// GetUpdatesChan returns a channel for receiving updates
func (b *Bot) GetUpdatesChan() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return b.api.GetUpdatesChan(u)
}

// SendMessage sends a text message
func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	return err
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

// EditMessage edits an existing message
func (b *Bot) EditMessage(chatID int64, messageID int, text string) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.api.Send(edit)
	return err
}

// EditMessageWithKeyboard edits an existing message and adds a keyboard
func (b *Bot) EditMessageWithKeyboard(chatID int64, messageID int, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard
	_, err := b.api.Send(edit)
	return err
}

// AnswerCallbackQuery answers a callback query
func (b *Bot) AnswerCallbackQuery(callbackID string, text string) error {
	callback := tgbotapi.NewCallback(callbackID, text)
	_, err := b.api.Send(callback)
	return err
}

// SetupCommands configures the bot commands with BotFather
func (b *Bot) SetupCommands() error {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "🏠 Welcome message and main menu",
		},
		{
			Command:     "menu",
			Description: "📋 Show main menu with all options",
		},
		{
			Command:     "learn",
			Description: "📚 Start learning session",
		},
		{
			Command:     "stats",
			Description: "📊 View your learning progress",
		},
		{
			Command:     "help",
			Description: "❓ Get help and instructions",
		},
	}

	setCommands := tgbotapi.NewSetMyCommands(commands...)
	_, err := b.api.Request(setCommands)
	if err != nil {
		return fmt.Errorf("failed to set bot commands: %w", err)
	}

	log.Printf("Bot commands configured successfully")
	return nil
}

// GetAPI returns the underlying bot API (for advanced usage)
func (b *Bot) GetAPI() *tgbotapi.BotAPI {
	return b.api
}
