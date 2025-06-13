package handlers

import (
	"context"
	"log"

	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/interfaces/telegram/handlers/shared"
)

// handleStatsFlow handles showing stats for both commands and callbacks
func (h *BotHandler) handleStatsFlow(ctx context.Context, chatID int64, messageID int, user *user.User, isCallback bool) {
	stats, err := h.learningUseCase.GetUserStats(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get user stats: %v", err)
		if isCallback {
			h.bot.EditMessage(chatID, messageID, "Sorry, there was an error getting your statistics.")
		} else {
			h.bot.SendMessage(chatID, "Sorry, there was an error getting your statistics.")
		}
		return
	}

	statsText := shared.FormatStatsText(stats)
	keyboard := shared.CreateStatsKeyboard(isCallback)

	if isCallback {
		h.bot.EditMessageWithKeyboard(chatID, messageID, statsText, keyboard)
	} else {
		h.bot.SendMessageWithKeyboard(chatID, statsText, keyboard)
	}
}

// handleHelpFlow handles showing help for both commands and callbacks
func (h *BotHandler) handleHelpFlow(ctx context.Context, chatID int64, messageID int, user *user.User, isCallback bool) {
	keyboard := shared.CreateHelpKeyboard(isCallback)
	if isCallback {
		h.bot.EditMessageWithKeyboard(chatID, messageID, shared.GetHelpText(), keyboard)
	} else {
		h.bot.SendMessageWithKeyboard(chatID, shared.GetHelpText(), keyboard)
	}
}

// handleLearningFlow handles starting learning for both commands and callbacks
func (h *BotHandler) handleLearningFlow(ctx context.Context, chatID int64, messageID int, user *user.User, isCallback bool) {
	session, err := h.learningUseCase.GetNextDueWord(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get next due word: %v", err)
		if isCallback {
			h.bot.EditMessage(chatID, messageID, "Sorry, there was an error getting your words. Please try again.")
		} else {
			h.bot.SendMessage(chatID, "Sorry, there was an error getting your words. Please try again.")
		}
		return
	}

	if session == nil {
		noWordsText := "🎉 Great job! You have no words due for review right now. Check back later!"
		keyboard := shared.CreateNoWordsKeyboard()

		if isCallback {
			h.bot.EditMessageWithKeyboard(chatID, messageID, noWordsText, keyboard)
		} else {
			h.bot.SendMessageWithKeyboard(chatID, noWordsText, keyboard)
		}
		return
	}

	// Store the session
	h.activeSessions[int64(user.ID())] = session

	// Send question
	if isCallback {
		h.sendQuestionAsEdit(chatID, messageID, session)
	} else {
		h.sendQuestion(chatID, session)
	}
}
