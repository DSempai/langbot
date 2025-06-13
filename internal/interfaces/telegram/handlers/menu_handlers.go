package handlers

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/interfaces/telegram/handlers/shared"
)

// handleMenuSelection processes menu button selections
func (h *BotHandler) handleMenuSelection(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User, selection string) {
	log.Printf("Menu selection: %s", selection)
	switch selection {
	case "menu_learn":
		h.handleMenuLearn(ctx, callback, user)
	case "menu_stats":
		h.handleMenuStats(ctx, callback, user)
	case "menu_help":
		h.handleMenuHelp(ctx, callback, user)
	case "menu_settings":
		h.handleMenuSettings(ctx, callback, user)
	default:
		log.Printf("Unknown menu selection: %s", selection)
	}
}

// handleBackToMenu returns to the main menu
func (h *BotHandler) handleBackToMenu(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	menuText := "🇳🇱 **Dutch Learning Bot - Main Menu**\n\nChoose an option:"
	h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, menuText, shared.CreateMainMenuKeyboard())
}

// handleMenuLearn starts learning from menu
func (h *BotHandler) handleMenuLearn(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	h.handleLearningFlow(ctx, callback.Message.Chat.ID, callback.Message.MessageID, user, true)
}

// handleMenuStats shows stats from menu
func (h *BotHandler) handleMenuStats(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	h.handleStatsFlow(ctx, callback.Message.Chat.ID, callback.Message.MessageID, user, true)
}

// handleMenuHelp shows help from menu
func (h *BotHandler) handleMenuHelp(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	h.handleHelpFlow(ctx, callback.Message.Chat.ID, callback.Message.MessageID, user, true)
}

// handleMenuSettings shows settings from menu
func (h *BotHandler) handleMenuSettings(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	// Get user preferences
	prefs, err := h.userUseCase.GetUserPreferences(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get user preferences: %v", err)
		h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			"Sorry, there was an error loading your settings. Please try again.")
		return
	}

	// Get current settings status
	grammarTipsStatus := "❌ **DISABLED**"
	grammarTipsAction := "Enable"
	if prefs.GrammarTipsEnabled() {
		grammarTipsStatus = "✅ **ENABLED**"
		grammarTipsAction = "Disable"
	}

	smartRemindersStatus := "❌ **DISABLED**"
	smartRemindersAction := "Enable"
	if prefs.SmartRemindersEnabled() {
		smartRemindersStatus = "✅ **ENABLED**"
		smartRemindersAction = "Disable"
	}

	reminderInterval := prefs.GetReminderInterval()

	// Build settings message
	settingsText := fmt.Sprintf(
		"⚙️ **Settings**\n\n"+
			"🔤 Grammar Tips: %s\n"+
			"⏰ Smart Reminders: %s\n"+
			"⌛️ Reminder Interval: **%d minutes**\n\n"+
			"_Use the buttons below to adjust settings:_",
		grammarTipsStatus, smartRemindersStatus, reminderInterval)

	// Create settings keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🔤 %s Grammar Tips", grammarTipsAction),
				"toggle_grammar_tips"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("⏰ %s Smart Reminders", smartRemindersAction),
				"toggle_smart_reminders"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➖ 15min", "set_interval_minus-15"),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("⏰ %dmin", reminderInterval), "noop"),
			tgbotapi.NewInlineKeyboardButtonData("➕ 15min", "set_interval_plus-15"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
		),
	)

	h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, settingsText, keyboard)
}
