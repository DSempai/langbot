package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dutch-learning-bot/internal/application/usecases"
	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/infrastructure/telegram"
)

// BotHandler handles Telegram bot interactions
type BotHandler struct {
	bot             *telegram.Bot
	userUseCase     *usecases.UserUseCase
	learningUseCase *usecases.LearningUseCase
	preferencesRepo user.PreferencesRepository
	activeSessions  map[int64]*usecases.LearningSession
}

// NewBotHandler creates a new bot handler
func NewBotHandler(
	bot *telegram.Bot,
	userUseCase *usecases.UserUseCase,
	learningUseCase *usecases.LearningUseCase,
	preferencesRepo user.PreferencesRepository,
) *BotHandler {
	return &BotHandler{
		bot:             bot,
		userUseCase:     userUseCase,
		learningUseCase: learningUseCase,
		preferencesRepo: preferencesRepo,
		activeSessions:  make(map[int64]*usecases.LearningSession),
	}
}

// Start starts the bot and handles updates
func (h *BotHandler) Start(ctx context.Context) error {
	updates := h.bot.GetUpdatesChan()

	log.Println("Bot started. Waiting for updates...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Bot stopping...")
			return nil
		case update := <-updates:
			go h.handleUpdate(update)
		}
	}
}

// handleUpdate processes incoming updates
func (h *BotHandler) handleUpdate(update tgbotapi.Update) {
	ctx := context.Background()

	if update.Message != nil {
		h.handleMessage(ctx, update.Message)
	} else if update.CallbackQuery != nil {
		h.handleCallbackQuery(ctx, update.CallbackQuery)
	}
}

// handleMessage processes text messages and commands
func (h *BotHandler) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	user, err := h.getOrCreateUser(ctx, message.From)
	if err != nil {
		log.Printf("Failed to get/create user: %v", err)
		return
	}

	switch message.Command() {
	case "start":
		h.handleStart(ctx, message, user)
	case "menu":
		h.handleMenu(ctx, message, user)
	case "learn":
		h.handleLearn(ctx, message, user)
	case "stats":
		h.handleStats(ctx, message, user)
	case "help":
		h.handleHelp(ctx, message, user)
	case "settings":
		// Redirect /settings command to menu settings
		h.handleMenuSettings(ctx, &tgbotapi.CallbackQuery{
			Message: message,
			From:    message.From,
		}, user)
	default:
		h.bot.SendMessage(message.Chat.ID, "Use /menu to see available options, or /help for detailed help.")
	}
}

// handleCallbackQuery processes inline keyboard callbacks
func (h *BotHandler) handleCallbackQuery(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	user, err := h.getOrCreateUser(ctx, callback.From)
	if err != nil {
		log.Printf("Failed to get/create user: %v", err)
		return
	}

	// Answer the callback to remove loading state
	if err := h.bot.AnswerCallbackQuery(callback.ID, ""); err != nil {
		log.Printf("Failed to answer callback query: %v", err)
	}

	data := callback.Data
	parts := strings.Split(data, "_")

	log.Printf("Processing callback: data=%s, parts=%v, message_id=%d", data, parts, callback.Message.MessageID)

	if len(parts) < 1 {
		log.Printf("Invalid callback data format: %s", data)
		return
	}

	// Handle noop callback (do nothing)
	if data == "noop" {
		return
	}

	switch parts[0] {
	case "menu":
		if len(parts) >= 2 {
			log.Printf("Handling menu selection: %s", data)
			h.handleMenuSelection(ctx, callback, user, data)
		} else {
			log.Printf("Invalid menu callback format: %s", data)
		}
	case "choice":
		if len(parts) >= 2 {
			h.handleMultipleChoice(ctx, callback, user, parts[1])
		}
	case "rating":
		if len(parts) >= 2 {
			h.handleRating(ctx, callback, user, parts[1])
		}
	case "continue":
		if len(parts) >= 2 && parts[1] == "learning" {
			h.handleContinueLearning(ctx, callback, user)
		}
	case "view":
		if len(parts) >= 2 && parts[1] == "stats" {
			h.handleViewStats(ctx, callback, user)
		}
	case "finish":
		if len(parts) >= 2 && parts[1] == "session" {
			h.handleFinishSession(ctx, callback, user)
		}
	case "back":
		if len(parts) >= 2 && parts[1] == "menu" {
			h.handleBackToMenu(ctx, callback, user)
		}
	case "toggle":
		if len(parts) >= 2 {
			// Join the remaining parts with underscore to handle multi-part identifiers
			identifier := strings.Join(parts[1:], "_")
			switch identifier {
			case "grammar_tips":
				h.handleToggleGrammarTips(ctx, callback, user)
			case "smart_reminders":
				h.handleToggleSmartReminders(ctx, callback, user)
			}
		}
	case "set":
		if len(parts) >= 3 && parts[1] == "interval" {
			// Split the last part by hyphen to get the direction and amount
			intervalParts := strings.Split(parts[2], "-")
			if len(intervalParts) == 2 && intervalParts[1] == "15" {
				switch intervalParts[0] {
				case "minus":
					h.handleAdjustInterval(ctx, callback, user, -15)
				case "plus":
					h.handleAdjustInterval(ctx, callback, user, 15)
				}
			}
		}
	default:
		log.Printf("Unknown callback type: %s", parts[0])
	}
}

// getOrCreateUser gets or creates a user from Telegram user info
func (h *BotHandler) getOrCreateUser(ctx context.Context, from *tgbotapi.User) (*user.User, error) {
	telegramID := user.TelegramID(from.ID)
	username := from.UserName
	firstName := from.FirstName
	lastName := from.LastName
	languageCode := from.LanguageCode

	return h.userUseCase.GetOrCreateUser(ctx, telegramID, username, firstName, lastName, languageCode)
}

// handleAdjustInterval adjusts the reminder interval by the specified amount
func (h *BotHandler) handleAdjustInterval(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User, adjustment int) {
	// Get current preferences
	prefs, err := h.userUseCase.GetUserPreferences(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get user preferences: %v", err)
		return
	}

	// Calculate new interval
	currentInterval := prefs.GetReminderInterval()
	newInterval := currentInterval + adjustment

	// Ensure minimum interval of 1 minute
	if newInterval < 1 {
		newInterval = 1
	}

	// Update the interval
	prefs.SetReminderInterval(newInterval)
	if err := h.userUseCase.UpdateUserPreferences(ctx, prefs); err != nil {
		log.Printf("Failed to update reminder interval: %v", err)
		return
	}

	// Get updated preferences to ensure we have the latest state
	prefs, err = h.userUseCase.GetUserPreferences(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get updated preferences: %v", err)
		return
	}

	// Construct new message with updated state
	message := fmt.Sprintf("*Settings*\n\n"+
		"Grammar Tips: %s\n"+
		"Smart Reminders: %s\n"+
		"Reminder Interval: %d minutes\n\n"+
		"Use the buttons below to adjust your settings.",
		getToggleEmoji(prefs.GrammarTipsEnabled()),
		getToggleEmoji(prefs.SmartRemindersEnabled()),
		prefs.GetReminderInterval())

	// Create keyboard with updated state
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("Grammar Tips %s", getToggleEmoji(prefs.GrammarTipsEnabled())),
				"toggle_grammar_tips",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("Smart Reminders %s", getToggleEmoji(prefs.SmartRemindersEnabled())),
				"toggle_smart_reminders",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏪ -15", "set_interval_minus-15"),
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("⏱ %d min", prefs.GetReminderInterval()), "noop"),
			tgbotapi.NewInlineKeyboardButtonData("+15 ⏩", "set_interval_plus-15"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("« Back to Menu", "back_menu"),
		),
	)

	// Edit the message with new content and keyboard
	if err := h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, message, keyboard); err != nil {
		log.Printf("Failed to update settings message: %v", err)
	}
}

// handleToggleGrammarTips handles toggling grammar tips
func (h *BotHandler) handleToggleGrammarTips(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	// Toggle the setting using the dedicated method
	_, err := h.userUseCase.ToggleGrammarTips(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to toggle grammar tips: %v", err)
		h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			"Sorry, there was an error updating your settings. Please try again.")
		return
	}

	// Show updated settings
	h.handleMenuSettings(ctx, callback, user)
}

// handleToggleSmartReminders handles toggling smart reminders
func (h *BotHandler) handleToggleSmartReminders(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	// Toggle the setting using the dedicated method
	_, err := h.userUseCase.ToggleSmartReminders(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to toggle smart reminders: %v", err)
		h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			"Sorry, there was an error updating your settings. Please try again.")
		return
	}

	// Show updated settings
	h.handleMenuSettings(ctx, callback, user)
}

// getToggleEmoji returns the appropriate emoji for a toggle state
func getToggleEmoji(enabled bool) string {
	if enabled {
		return "✅"
	}
	return "❌"
}
