package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dutch-learning-bot/internal/application/usecases"
	"dutch-learning-bot/internal/domain/learning"
	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/interfaces/telegram/handlers/shared"
)

// clickTracker tracks recent clicks to prevent rapid duplicates
type clickTracker struct {
	mu          sync.RWMutex
	lastClicks  map[string]time.Time
	cleanupTick *time.Ticker
}

// newClickTracker creates a new click tracker
func newClickTracker() *clickTracker {
	ct := &clickTracker{
		lastClicks:  make(map[string]time.Time),
		cleanupTick: time.NewTicker(30 * time.Second),
	}

	// Clean up old entries every 30 seconds
	go func() {
		for range ct.cleanupTick.C {
			ct.cleanup()
		}
	}()

	return ct
}

// isRecentClick checks if this click is too recent (debouncing)
func (ct *clickTracker) isRecentClick(userID int64, action string) bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	key := fmt.Sprintf("%d_%s", userID, action)
	lastClick, exists := ct.lastClicks[key]
	if !exists {
		return false
	}

	// Prevent clicks within 1 second for same action
	return time.Since(lastClick) < time.Second
}

// recordClick records a click timestamp
func (ct *clickTracker) recordClick(userID int64, action string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	key := fmt.Sprintf("%d_%s", userID, action)
	ct.lastClicks[key] = time.Now()
}

// cleanup removes old click records
func (ct *clickTracker) cleanup() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	for key, timestamp := range ct.lastClicks {
		if timestamp.Before(cutoff) {
			delete(ct.lastClicks, key)
		}
	}
}

// Global click tracker instance
var globalClickTracker = newClickTracker()

// sendQuestion sends a learning question to the user
func (h *BotHandler) sendQuestion(chatID int64, session *usecases.LearningSession) {
	var questionText string
	var hintText string

	if session.QuestionType == usecases.QuestionTypeEnglishToDutch {
		questionText = fmt.Sprintf("🇬🇧➡️🇳🇱 Translate to Dutch:\n\n**%s**", session.Word.English())
		hintText = fmt.Sprintf("Category: %s", session.Word.Category())
	} else {
		questionText = fmt.Sprintf("🇳🇱➡️🇬🇧 Translate to English:\n\n**%s**", session.Word.Dutch())
		hintText = fmt.Sprintf("Category: %s", session.Word.Category())
	}

	fullText := fmt.Sprintf("%s\n\n💡 %s", questionText, hintText)

	// Add grammar tip if present (surprise feature!)
	if session.GrammarTip != nil {
		fullText += fmt.Sprintf("\n\n🎯 **Grammar Tip: %s**\n%s",
			session.GrammarTip.Title(),
			session.GrammarTip.Explanation())

		// Add an example if available
		if len(session.GrammarTip.DutchExample()) > 0 || len(session.GrammarTip.EnglishExample()) > 0 {
			fullText += fmt.Sprintf("\n\n🇳🇱 %s\n🇬🇧 %s", session.GrammarTip.DutchExample(), session.GrammarTip.EnglishExample())
		}
	}

	fullText += "\n\nChoose the correct translation:"

	// Create multiple choice keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("A) "+session.Options[0], "choice_0"),
			tgbotapi.NewInlineKeyboardButtonData("B) "+session.Options[1], "choice_1"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("C) "+session.Options[2], "choice_2"),
			tgbotapi.NewInlineKeyboardButtonData("D) "+session.Options[3], "choice_3"),
		),
	)

	h.bot.SendMessageWithKeyboard(chatID, fullText, keyboard)
}

// sendQuestionAsEdit sends a learning question by editing an existing message
func (h *BotHandler) sendQuestionAsEdit(chatID int64, messageID int, session *usecases.LearningSession) {
	var questionText string
	var hintText string

	if session.QuestionType == usecases.QuestionTypeEnglishToDutch {
		questionText = fmt.Sprintf("🇬🇧➡️🇳🇱 Translate to Dutch:\n\n*%s*", shared.EscapeMarkdown(session.Word.English()))
		hintText = fmt.Sprintf("Category: %s", shared.EscapeMarkdown(string(session.Word.Category())))
	} else {
		questionText = fmt.Sprintf("🇳🇱➡️🇬🇧 Translate to English:\n\n*%s*", shared.EscapeMarkdown(session.Word.Dutch()))
		hintText = fmt.Sprintf("Category: %s", shared.EscapeMarkdown(string(session.Word.Category())))
	}

	fullText := fmt.Sprintf("%s\n\n💡 %s", questionText, hintText)

	// Add grammar tip if present (surprise feature!)
	if session.GrammarTip != nil {
		fullText += fmt.Sprintf("\n\n🎯 *Grammar Tip: %s*\n%s",
			shared.EscapeMarkdown(session.GrammarTip.Title()),
			shared.EscapeMarkdown(session.GrammarTip.Explanation()))

		// Add an example if available
		if len(session.GrammarTip.DutchExample()) > 0 || len(session.GrammarTip.EnglishExample()) > 0 {
			fullText += fmt.Sprintf("\n\n🇳🇱 %s\n🇬🇧 %s",
				shared.EscapeMarkdown(session.GrammarTip.DutchExample()),
				shared.EscapeMarkdown(session.GrammarTip.EnglishExample()))
		}
	}

	fullText += "\n\nChoose the correct translation:"

	// Create multiple choice keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("A) "+shared.EscapeMarkdown(session.Options[0]), "choice_0"),
			tgbotapi.NewInlineKeyboardButtonData("B) "+shared.EscapeMarkdown(session.Options[1]), "choice_1"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("C) "+shared.EscapeMarkdown(session.Options[2]), "choice_2"),
			tgbotapi.NewInlineKeyboardButtonData("D) "+shared.EscapeMarkdown(session.Options[3]), "choice_3"),
		),
	)

	log.Printf("Sending question: %s", fullText)
	err := h.bot.EditMessageWithKeyboard(chatID, messageID, fullText, keyboard)
	if err != nil {
		log.Printf("Failed to send question: %v", err)
		// Try to send error message
		h.bot.EditMessage(chatID, messageID, "Sorry, there was an error displaying the question. Please try again with /learn")
	}
}

// handleMultipleChoice processes multiple choice selection
func (h *BotHandler) handleMultipleChoice(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User, choiceStr string) {
	// Debounce rapid clicks
	userID := int64(user.ID())
	if globalClickTracker.isRecentClick(userID, "choice_"+choiceStr) {
		log.Printf("Ignoring rapid duplicate click from user %d for choice %s", userID, choiceStr)
		return
	}
	globalClickTracker.recordClick(userID, "choice_"+choiceStr)

	session, exists := h.activeSessions[userID]
	if !exists {
		h.bot.SendMessage(callback.Message.Chat.ID, "No active session found. Use /learn to start.")
		return
	}

	choiceIndex, err := strconv.Atoi(choiceStr)
	if err != nil {
		log.Printf("Invalid choice index: %s", choiceStr)
		return
	}

	// Check if the answer is correct
	isCorrect := h.learningUseCase.CheckMultipleChoiceAnswer(session, choiceIndex)

	// Show result
	var resultText string
	selectedAnswer := session.Options[choiceIndex]
	correctAnswer := session.Options[session.CorrectIndex]

	if isCorrect {
		resultText = fmt.Sprintf("✅ **Correct!**\n\nYour answer: %s\n\n🇬🇧 %s\n🇳🇱 %s",
			selectedAnswer, session.Word.English(), session.Word.Dutch())
	} else {
		resultText = fmt.Sprintf("❌ **Incorrect**\n\nYour answer: %s\nCorrect answer: %s\n\n🇬🇧 %s\n🇳🇱 %s",
			selectedAnswer, correctAnswer, session.Word.English(), session.Word.Dutch())
	}

	// Add rating request
	resultText += "\n\nHow well did you know this word?"

	// Create rating keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("😵 Again", "rating_1"),
			tgbotapi.NewInlineKeyboardButtonData("😐 Hard", "rating_2"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🙂 Good", "rating_3"),
			tgbotapi.NewInlineKeyboardButtonData("😄 Easy", "rating_4"),
		),
	)

	// Edit the original message
	h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, resultText, keyboard)
}

// handleRating processes rating selection
func (h *BotHandler) handleRating(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User, ratingStr string) {
	userID := int64(user.ID())

	// Debounce rapid clicks
	if globalClickTracker.isRecentClick(userID, "rating_"+ratingStr) {
		log.Printf("Ignoring rapid duplicate rating click from user %d for rating %s", userID, ratingStr)
		return
	}
	globalClickTracker.recordClick(userID, "rating_"+ratingStr)

	session, exists := h.activeSessions[userID]
	if !exists {
		h.bot.SendMessage(callback.Message.Chat.ID, "No active session found. Use /learn to start.")
		return
	}

	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		log.Printf("Invalid rating: %s", ratingStr)
		return
	}

	// Process in the background to improve responsiveness
	go func() {
		// Create a timeout context for this operation
		bgCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Calculate response time
		responseTime := time.Since(session.StartTime)

		// Process the review
		err := h.learningUseCase.ProcessReview(bgCtx, session, learning.Rating(rating), responseTime)
		if err != nil {
			log.Printf("Failed to process review: %v", err)
			h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
				"❌ Error processing review. Please try again with /learn")
			return
		}

		// Clean up current session
		delete(h.activeSessions, userID)

		// Get the next word
		nextSession, err := h.learningUseCase.GetNextDueWord(bgCtx, user.ID())
		if err != nil {
			log.Printf("Failed to get next word: %v", err)
			h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
				"❌ Error getting next word. Please try again with /learn")
			return
		}

		if nextSession != nil {
			// Store the new session
			h.activeSessions[userID] = nextSession
			// Show the next question
			h.sendQuestionAsEdit(callback.Message.Chat.ID, callback.Message.MessageID, nextSession)
		} else {
			// No more words to review
			resultText := "🎉 Great job! You have no more words due for review right now."
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("📊 View Stats", "menu_stats"),
					tgbotapi.NewInlineKeyboardButtonData("🏠 Main Menu", "back_menu"),
				),
			)
			h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, resultText, keyboard)
		}
	}()
}

// handleViewStats shows user statistics
func (h *BotHandler) handleViewStats(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	h.handleStatsFlow(ctx, callback.Message.Chat.ID, callback.Message.MessageID, user, true)
}

// handleContinueLearning handles the continue learning button
func (h *BotHandler) handleContinueLearning(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	h.handleLearningFlow(ctx, callback.Message.Chat.ID, callback.Message.MessageID, user, true)
}

// handleFinishSession handles the finish session button
func (h *BotHandler) handleFinishSession(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	// Clean up session
	delete(h.activeSessions, int64(user.ID()))

	// Show main menu
	h.handleBackToMenu(ctx, callback, user)
}
