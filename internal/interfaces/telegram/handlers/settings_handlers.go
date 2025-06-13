package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/interfaces/telegram"
)

// SettingsHandlers handles settings-related commands
type SettingsHandlers struct {
	bot             *tgbotapi.BotAPI
	preferencesRepo user.PreferencesRepository
}

// NewSettingsHandlers creates a new settings handlers instance
func NewSettingsHandlers(
	bot *tgbotapi.BotAPI,
	preferencesRepo user.PreferencesRepository,
) *SettingsHandlers {
	return &SettingsHandlers{
		bot:             bot,
		preferencesRepo: preferencesRepo,
	}
}

const (
	// Command constants
	cmdSettings             = "settings"
	cmdToggleGrammarTips    = "toggle_grammar_tips"
	cmdToggleSmartReminders = "toggle_smart_reminders"
	cmdSetReminderInterval  = "set_reminder_interval"
)

// RegisterHandlers registers all settings-related handlers
func (h *SettingsHandlers) RegisterHandlers(dispatcher telegram.Dispatcher) {
	dispatcher.RegisterHandler(cmdSettings, h.handleSettings)
	dispatcher.RegisterHandler(cmdToggleGrammarTips, h.handleToggleGrammarTips)
	dispatcher.RegisterHandler(cmdToggleSmartReminders, h.handleToggleSmartReminders)
	dispatcher.RegisterHandler(cmdSetReminderInterval, h.handleSetReminderInterval)
}

// handleToggleGrammarTips handles the toggle_grammar_tips command
func (h *SettingsHandlers) handleToggleGrammarTips(ctx context.Context, update tgbotapi.Update) error {
	userID := user.ID(update.Message.From.ID)
	preferences, err := h.preferencesRepo.FindPreferences(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}

	// Toggle the setting
	preferences.SetGrammarTipsEnabled(!preferences.GrammarTipsEnabled())

	// Save the updated preferences
	if err := h.preferencesRepo.SavePreferences(ctx, preferences); err != nil {
		return fmt.Errorf("failed to save preferences: %w", err)
	}

	// Send confirmation
	status := "enabled"
	if !preferences.GrammarTipsEnabled() {
		status = "disabled"
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Grammar tips %s", status))
	_, err = h.bot.Send(msg)
	return err
}

// handleToggleSmartReminders handles the toggle_smart_reminders command
func (h *SettingsHandlers) handleToggleSmartReminders(ctx context.Context, update tgbotapi.Update) error {
	userID := user.ID(update.Message.From.ID)
	preferences, err := h.preferencesRepo.FindPreferences(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}

	// Toggle the setting
	preferences.SetSmartRemindersEnabled(!preferences.SmartRemindersEnabled())

	// Save the updated preferences
	if err := h.preferencesRepo.SavePreferences(ctx, preferences); err != nil {
		return fmt.Errorf("failed to save preferences: %w", err)
	}

	// Send confirmation
	status := "enabled"
	if !preferences.SmartRemindersEnabled() {
		status = "disabled"
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Smart reminders %s", status))
	_, err = h.bot.Send(msg)
	return err
}

// handleSettings handles the settings command
func (h *SettingsHandlers) handleSettings(ctx context.Context, update tgbotapi.Update) error {
	userID := user.ID(update.Message.From.ID)
	preferences, err := h.preferencesRepo.FindPreferences(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}

	// Build settings message
	var sb strings.Builder
	sb.WriteString("⚙️ *Settings*\n\n")

	// Grammar tips status
	grammarTipsStatus := "✅ *ENABLED*"
	if !preferences.GrammarTipsEnabled() {
		grammarTipsStatus = "❌ *DISABLED*"
	}
	sb.WriteString(fmt.Sprintf("Grammar Tips: %s\n", grammarTipsStatus))

	// Smart reminders status
	smartRemindersStatus := "✅ *ENABLED*"
	if !preferences.SmartRemindersEnabled() {
		smartRemindersStatus = "❌ *DISABLED*"
	}
	sb.WriteString(fmt.Sprintf("Smart Reminders: %s\n", smartRemindersStatus))

	// Reminder interval
	sb.WriteString(fmt.Sprintf("Reminder Interval: *%d minutes*\n", preferences.GetReminderInterval()))

	// Add commands
	sb.WriteString("\n*Commands:*\n")
	sb.WriteString("/toggle\\_grammar\\_tips - Toggle grammar tips\n")
	sb.WriteString("/toggle\\_smart\\_reminders - Toggle smart reminders\n")
	sb.WriteString("/set\\_reminder\\_interval <minutes> - Set reminder interval\n")

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	_, err = h.bot.Send(msg)
	return err
}

// handleSetReminderInterval handles the set_reminder_interval command
func (h *SettingsHandlers) handleSetReminderInterval(ctx context.Context, update tgbotapi.Update) error {
	args := strings.Fields(update.Message.Text)
	if len(args) != 2 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Please specify the reminder interval in minutes.\n"+
				"Example: /set\\_reminder\\_interval 30")
		msg.ParseMode = "Markdown"
		_, err := h.bot.Send(msg)
		return err
	}

	// Parse interval
	interval, err := strconv.Atoi(args[1])
	if err != nil || interval < 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Please provide a valid number of minutes (minimum 1).\n"+
				"Example: /set\\_reminder\\_interval 30")
		msg.ParseMode = "Markdown"
		_, err := h.bot.Send(msg)
		return err
	}

	// Get and update preferences
	userID := user.ID(update.Message.From.ID)
	preferences, err := h.preferencesRepo.FindPreferences(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}

	preferences.SetReminderInterval(interval)
	if err := h.preferencesRepo.SavePreferences(ctx, preferences); err != nil {
		return fmt.Errorf("failed to save preferences: %w", err)
	}

	// Send confirmation
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("✅ Reminder interval set to *%d minutes*", interval))
	msg.ParseMode = "Markdown"
	_, err = h.bot.Send(msg)
	return err
}
