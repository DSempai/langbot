package handlers

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/interfaces/telegram/handlers/shared"
)

// handleStart processes the /start command
func (h *BotHandler) handleStart(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	welcomeText := fmt.Sprintf(
		"🇳🇱 Welcome to Dutch Learning Bot, %s!\n\n"+
			"I'll help you learn Dutch using spaced repetition (FSRS algorithm).\n\n"+
			"Choose an option below to get started:",
		user.FirstName())

	h.bot.SendMessageWithKeyboard(message.Chat.ID, welcomeText, shared.CreateMainMenuKeyboard())
}

// handleMenu processes the /menu command
func (h *BotHandler) handleMenu(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	menuText := "🇳🇱 **Dutch Learning Bot - Main Menu**\n\nChoose an option:"
	h.bot.SendMessageWithKeyboard(message.Chat.ID, menuText, shared.CreateMainMenuKeyboard())
}

// handleLearn processes the /learn command
func (h *BotHandler) handleLearn(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	h.handleLearningFlow(ctx, message.Chat.ID, message.MessageID, user, false)
}

// handleStats processes the /stats command
func (h *BotHandler) handleStats(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	h.handleStatsFlow(ctx, message.Chat.ID, message.MessageID, user, false)
}

// handleHelp processes the /help command
func (h *BotHandler) handleHelp(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	h.handleHelpFlow(ctx, message.Chat.ID, message.MessageID, user, false)
}
