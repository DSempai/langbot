package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"dutch-learning-bot/internal/application/usecases"
	"dutch-learning-bot/internal/domain/learning"
	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/infrastructure/telegram"
)

// BotHandler handles Telegram bot interactions
type BotHandler struct {
	bot             *telegram.Bot
	userUseCase     *usecases.UserUseCase
	learningUseCase *usecases.LearningUseCase
	activeSessions  map[int64]*usecases.LearningSession
}

// NewBotHandler creates a new bot handler
func NewBotHandler(
	bot *telegram.Bot,
	userUseCase *usecases.UserUseCase,
	learningUseCase *usecases.LearningUseCase,
) *BotHandler {
	return &BotHandler{
		bot:             bot,
		userUseCase:     userUseCase,
		learningUseCase: learningUseCase,
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

// handleMessage processes text messages
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
	h.bot.AnswerCallbackQuery(callback.ID, "")

	data := callback.Data
	parts := strings.Split(data, "_")

	if len(parts) < 1 {
		return
	}

	switch parts[0] {
	case "menu":
		if len(parts) >= 2 {
			h.handleMenuSelection(ctx, callback, user, parts[1])
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
		if len(parts) >= 3 && parts[1] == "grammar" && parts[2] == "tips" {
			h.handleToggleGrammarTips(ctx, callback, user)
		} else if len(parts) >= 3 && parts[1] == "smart" && parts[2] == "reminders" {
			h.handleToggleSmartReminders(ctx, callback, user)
		}
	}
}

// handleStart processes the /start command
func (h *BotHandler) handleStart(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	welcomeText := fmt.Sprintf(
		"🇳🇱 Welcome to Dutch Learning Bot, %s!\n\n"+
			"I'll help you learn Dutch using spaced repetition (FSRS algorithm).\n\n"+
			"Choose an option below to get started:",
		user.FirstName())

	// Create main menu keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Start Learning", "menu_learn"),
			tgbotapi.NewInlineKeyboardButtonData("📊 View Stats", "menu_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "menu_help"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "menu_settings"),
		),
	)

	h.bot.SendMessageWithKeyboard(message.Chat.ID, welcomeText, keyboard)
}

// handleMenu processes the /menu command
func (h *BotHandler) handleMenu(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	menuText := "🇳🇱 **Dutch Learning Bot - Main Menu**\n\nChoose an option:"

	// Create main menu keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Start Learning", "menu_learn"),
			tgbotapi.NewInlineKeyboardButtonData("📊 View Stats", "menu_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "menu_help"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "menu_settings"),
		),
	)

	h.bot.SendMessageWithKeyboard(message.Chat.ID, menuText, keyboard)
}

// handleLearn processes the /learn command
func (h *BotHandler) handleLearn(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	session, err := h.learningUseCase.GetNextDueWord(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get next due word: %v", err)
		h.bot.SendMessage(message.Chat.ID, "Sorry, there was an error getting your words. Please try again.")
		return
	}

	if session == nil {
		noWordsText := "🎉 Great job! You have no words due for review right now. Check back later!"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 View Stats", "menu_stats"),
				tgbotapi.NewInlineKeyboardButtonData("🏠 Main Menu", "back_menu"),
			),
		)

		h.bot.SendMessageWithKeyboard(message.Chat.ID, noWordsText, keyboard)
		return
	}

	// Store the session
	h.activeSessions[int64(user.ID())] = session

	// Send question
	h.sendQuestion(message.Chat.ID, session)
}

// handleStats processes the /stats command
func (h *BotHandler) handleStats(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	stats, err := h.learningUseCase.GetUserStats(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get user stats: %v", err)
		h.bot.SendMessage(message.Chat.ID, "Sorry, there was an error getting your statistics.")
		return
	}

	statsText := fmt.Sprintf(
		"📊 **Your Learning Stats**\n\n"+
			"📚 Total words: %d\n"+
			"🆕 New: %d\n"+
			"📖 Learning: %d\n"+
			"✅ Review: %d\n"+
			"⏰ Due now: %d\n\n"+
			"🎯 Average difficulty: %.1f/10\n"+
			"📈 Total reviews: %d\n"+
			"✅ Correct answers: %d\n\n"+
			"Keep up the great work! 🌟",
		stats.TotalWords, stats.NewWords, stats.LearningWords, stats.ReviewWords,
		stats.DueWords, stats.AvgDifficulty, stats.TotalReviews, stats.CorrectReviews)

	// Create menu keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Start Learning", "menu_learn"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Main Menu", "back_menu"),
		),
	)

	h.bot.SendMessageWithKeyboard(message.Chat.ID, statsText, keyboard)
}

// handleHelp processes the /help command
func (h *BotHandler) handleHelp(ctx context.Context, message *tgbotapi.Message, user *user.User) {
	helpText := `🇳🇱 **Dutch Learning Bot Help**

**Available Commands:**
/start - Show welcome message
/menu - Show main menu
/learn - Start learning session
/stats - View your progress
/help - Show this help

**How it works:**
This bot uses the FSRS (Free Spaced Repetition System) algorithm to optimize your learning schedule. Based on how well you remember each word, the bot will schedule future reviews at optimal intervals.

**Rating Guide:**
😵 **Again** - You didn't remember at all
😐 **Hard** - You remembered but it was difficult
🙂 **Good** - You remembered with some effort
😄 **Easy** - You remembered easily

**Tips:**
- Be honest with your ratings for best results
- Practice regularly for optimal retention
- Focus on understanding rather than just memorizing

Good luck with your Dutch learning! 🍀`

	// Create menu keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Main Menu", "back_menu"),
		),
	)

	h.bot.SendMessageWithKeyboard(message.Chat.ID, helpText, keyboard)
}

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

// handleRating processes rating callbacks
func (h *BotHandler) handleRating(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User, ratingStr string) {
	session, exists := h.activeSessions[int64(user.ID())]
	if !exists {
		h.bot.SendMessage(callback.Message.Chat.ID, "No active session found. Use /learn to start.")
		return
	}

	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		log.Printf("Invalid rating: %s", ratingStr)
		return
	}

	// Show the answer first
	var answerText string
	answerText = fmt.Sprintf("🔍 **Answer**\n\n🇬🇧 %s\n🇳🇱 %s\n\n", session.Word.English(), session.Word.Dutch())

	// Calculate response time
	responseTime := time.Since(session.StartTime)

	// Process the review
	err = h.learningUseCase.ProcessReview(ctx, session, learning.Rating(rating), responseTime)
	if err != nil {
		log.Printf("Failed to process review: %v", err)
		h.bot.SendMessage(callback.Message.Chat.ID, "Error processing review. Please try again.")
		return
	}

	// Calculate next review time
	nextReview := session.Progress.FSRSCard().DueDate()
	var nextReviewText string

	if session.Progress.FSRSCard().State() == learning.StateLearning || session.Progress.FSRSCard().State() == learning.StateRelearning {
		duration := time.Until(nextReview)
		if duration < time.Hour {
			nextReviewText = fmt.Sprintf("Next review: %d minutes", int(duration.Minutes()))
		} else {
			nextReviewText = fmt.Sprintf("Next review: %d hours", int(duration.Hours()))
		}
	} else {
		days := int(time.Until(nextReview).Hours() / 24)
		if days == 0 {
			nextReviewText = "Next review: Today"
		} else if days == 1 {
			nextReviewText = "Next review: Tomorrow"
		} else {
			nextReviewText = fmt.Sprintf("Next review: %d days", days)
		}
	}

	// Clean up current session first
	delete(h.activeSessions, int64(user.ID()))

	// Get the next word immediately
	nextSession, err := h.learningUseCase.GetNextDueWord(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get next due word: %v", err)
		// Fall back to showing the continue menu if there's an error
		feedbackText := fmt.Sprintf("%s📊 **Review Complete**\n\n%s\n\nError getting next word. Please try again.", answerText, nextReviewText)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📚 Continue Learning", "continue_learning"),
				tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
			),
		)

		h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, feedbackText, keyboard)
		return
	}

	if nextSession == nil {
		// No more words due - show completion message
		completionText := fmt.Sprintf("%s📊 **Review Complete**\n\n%s\n\n🎉 Great job! You have no more words due for review right now. Check back later!", answerText, nextReviewText)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 View Stats", "menu_stats"),
				tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
			),
		)

		h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, completionText, keyboard)
		return
	}

	// Store the new session
	h.activeSessions[int64(user.ID())] = nextSession

	// Show next question immediately
	h.sendQuestionAsEdit(callback.Message.Chat.ID, callback.Message.MessageID, nextSession)
}

// handleMultipleChoice processes multiple choice selection
func (h *BotHandler) handleMultipleChoice(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User, choiceStr string) {
	session, exists := h.activeSessions[int64(user.ID())]
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

// handleContinueLearning starts a new learning session
func (h *BotHandler) handleContinueLearning(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	session, err := h.learningUseCase.GetNextDueWord(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get next due word: %v", err)
		h.bot.SendMessage(callback.Message.Chat.ID, "Sorry, there was an error getting your words. Please try again.")
		return
	}

	if session == nil {
		h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			callback.Message.Text+"\n\n🎉 Great job! You have no more words due for review right now. Check back later!")
		return
	}

	// Store the session
	h.activeSessions[int64(user.ID())] = session

	// Send new question by editing the current message
	h.sendQuestionAsEdit(callback.Message.Chat.ID, callback.Message.MessageID, session)
}

// handleViewStats shows user statistics
func (h *BotHandler) handleViewStats(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	stats, err := h.learningUseCase.GetUserStats(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get user stats: %v", err)
		h.bot.SendMessage(callback.Message.Chat.ID, "Sorry, there was an error getting your statistics.")
		return
	}

	statsText := fmt.Sprintf(
		"📊 **Your Learning Stats**\n\n"+
			"📚 Total words: %d\n"+
			"🆕 New: %d\n"+
			"📖 Learning: %d\n"+
			"✅ Review: %d\n"+
			"⏰ Due now: %d\n\n"+
			"🎯 Average difficulty: %.1f/10\n"+
			"📈 Total reviews: %d\n"+
			"✅ Correct answers: %d\n\n"+
			"Keep up the great work! 🌟",
		stats.TotalWords, stats.NewWords, stats.LearningWords, stats.ReviewWords,
		stats.DueWords, stats.AvgDifficulty, stats.TotalReviews, stats.CorrectReviews)

	// Create back to continue keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Continue Learning", "continue_learning"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Finish", "finish_session"),
		),
	)

	h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, statsText, keyboard)
}

// handleFinishSession ends the learning session
func (h *BotHandler) handleFinishSession(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	finishText := fmt.Sprintf("✅ **Session Complete!**\n\n" +
		"Great work! You can start a new session anytime.\n\n" +
		"Use /menu to see all available options.")

	// Create back to menu keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
		),
	)

	h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, finishText, keyboard)
}

// handleMenuSelection processes menu button selections
func (h *BotHandler) handleMenuSelection(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User, selection string) {
	switch selection {
	case "learn":
		h.handleMenuLearn(ctx, callback, user)
	case "stats":
		h.handleMenuStats(ctx, callback, user)
	case "help":
		h.handleMenuHelp(ctx, callback, user)
	case "settings":
		h.handleMenuSettings(ctx, callback, user)
	}
}

// handleBackToMenu returns to the main menu
func (h *BotHandler) handleBackToMenu(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	menuText := "🇳🇱 **Dutch Learning Bot - Main Menu**\n\nChoose an option:"

	// Create main menu keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Start Learning", "menu_learn"),
			tgbotapi.NewInlineKeyboardButtonData("📊 View Stats", "menu_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "menu_help"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "menu_settings"),
		),
	)

	h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, menuText, keyboard)
}

// handleMenuLearn starts learning from menu
func (h *BotHandler) handleMenuLearn(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	session, err := h.learningUseCase.GetNextDueWord(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get next due word: %v", err)
		h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			"Sorry, there was an error getting your words. Please try again.")
		return
	}

	if session == nil {
		// No words due - show message with back to menu option
		noWordsText := "🎉 Great job! You have no words due for review right now. Check back later!"

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
			),
		)

		h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, noWordsText, keyboard)
		return
	}

	// Store the session
	h.activeSessions[int64(user.ID())] = session

	// Send question by editing the current message
	h.sendQuestionAsEdit(callback.Message.Chat.ID, callback.Message.MessageID, session)
}

// handleMenuStats shows stats from menu
func (h *BotHandler) handleMenuStats(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	stats, err := h.learningUseCase.GetUserStats(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get user stats: %v", err)
		h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			"Sorry, there was an error getting your statistics.")
		return
	}

	statsText := fmt.Sprintf(
		"📊 **Your Learning Stats**\n\n"+
			"📚 Total words: %d\n"+
			"🆕 New: %d\n"+
			"📖 Learning: %d\n"+
			"✅ Review: %d\n"+
			"⏰ Due now: %d\n\n"+
			"🎯 Average difficulty: %.1f/10\n"+
			"📈 Total reviews: %d\n"+
			"✅ Correct answers: %d\n\n"+
			"Keep up the great work! 🌟",
		stats.TotalWords, stats.NewWords, stats.LearningWords, stats.ReviewWords,
		stats.DueWords, stats.AvgDifficulty, stats.TotalReviews, stats.CorrectReviews)

	// Create back to menu keyboard with learn option
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Start Learning", "menu_learn"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
		),
	)

	h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, statsText, keyboard)
}

// handleMenuHelp shows help from menu
func (h *BotHandler) handleMenuHelp(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	helpText := `🇳🇱 **Dutch Learning Bot Help**

**Available Commands:**
/start - Show welcome message
/menu - Show main menu
/learn - Start learning session
/stats - View your progress
/help - Show this help

**How it works:**
This bot uses the FSRS (Free Spaced Repetition System) algorithm to optimize your learning schedule. Based on how well you remember each word, the bot will schedule future reviews at optimal intervals.

**Rating Guide:**
😵 **Again** - You didn't remember at all
😐 **Hard** - You remembered but it was difficult  
🙂 **Good** - You remembered with some effort
😄 **Easy** - You remembered easily

**Tips:**
- Be honest with your ratings for best results
- Practice regularly for optimal retention
- Focus on understanding rather than just memorizing

Good luck with your Dutch learning! 🍀`

	// Create back to menu keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Start Learning", "menu_learn"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
		),
	)

	h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, helpText, keyboard)
}

// handleMenuSettings shows settings from menu
func (h *BotHandler) handleMenuSettings(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	// Get user preferences
	preferences, err := h.userUseCase.GetUserPreferences(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to get user preferences: %v", err)
		h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			"Sorry, there was an error loading your settings. Please try again.")
		return
	}

	// Get current settings status
	grammarTipsStatus := "❌ **DISABLED**"
	grammarTipsAction := "Enable"
	if preferences.GrammarTipsEnabled() {
		grammarTipsStatus = "✅ **ENABLED**"
		grammarTipsAction = "Disable"
	}

	smartRemindersStatus := "❌ **DISABLED**"
	smartRemindersAction := "Enable"
	if preferences.SmartRemindersEnabled() {
		smartRemindersStatus = "✅ **ENABLED**"
		smartRemindersAction = "Disable"
	}

	settingsText := fmt.Sprintf(`⚙️ **Settings**

**Grammar Tips (Pro Tips)** 🎯
%s

**Smart Reminders** 🔔
%s

**How Grammar Tips work:**
• **Contextual & Smart**: Tips appear only when relevant to the word you're learning
• **20%% Chance**: Tips show with 20%% probability during learning sessions
• **Category-Based**: Each tip is designed for specific word categories (home, body, verbs, etc.)
• **Pattern Recognition**: Tips trigger based on word patterns (e.g., words ending in -kamer, compound words)

**Grammar Tip Examples:**
• Learning "slaapkamer"? → Get tip about compound words with -kamer
• Learning "koelkast"? → Get tip about compound appliances
• Learning "werk"? → Get tip about Dutch verb past tense formation
• Learning body parts? → Get tip about articles with body parts

**Smart Reminders:**
• Checks every 30 minutes for users with due words
• Only reminds during active hours (8 AM - 10 PM)
• Max 3 reminders per day (won't spam you!)
• Smart timing based on your activity patterns

**Current Configuration:**
🎯 Target retention: 90%%
📊 FSRS algorithm: v4
🔀 Question types: Both directions
📝 Answer format: Multiple choice

**Available Features:**
✅ Spaced repetition learning
✅ Progress tracking
✅ Performance statistics
✅ Anti-repetition system
✅ Contextual grammar guidance`, grammarTipsStatus, smartRemindersStatus)

	// Create settings keyboard with toggle options
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 "+grammarTipsAction+" Grammar Tips", "toggle_grammar_tips"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔔 "+smartRemindersAction+" Smart Reminders", "toggle_smart_reminders"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Start Learning", "menu_learn"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
		),
	)

	h.bot.EditMessageWithKeyboard(callback.Message.Chat.ID, callback.Message.MessageID, settingsText, keyboard)
}

// handleToggleGrammarTips toggles the grammar tips setting
func (h *BotHandler) handleToggleGrammarTips(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	// Toggle the setting
	newState, err := h.userUseCase.ToggleGrammarTips(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to toggle grammar tips setting: %v", err)
		h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			"Sorry, there was an error updating your settings. Please try again.")
		return
	}

	// Show confirmation and return to settings
	status := "disabled"
	emoji := "❌"
	if newState {
		status = "enabled"
		emoji = "✅"
	}

	confirmText := fmt.Sprintf("%s Grammar tips %s! Returning to settings...", emoji, status)
	h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID, confirmText)

	// Wait briefly, then show settings again
	time.Sleep(1 * time.Second)
	h.handleMenuSettings(ctx, callback, user)
}

// handleToggleSmartReminders toggles the smart reminders setting
func (h *BotHandler) handleToggleSmartReminders(ctx context.Context, callback *tgbotapi.CallbackQuery, user *user.User) {
	// Toggle the setting
	newState, err := h.userUseCase.ToggleSmartReminders(ctx, user.ID())
	if err != nil {
		log.Printf("Failed to toggle smart reminders setting: %v", err)
		h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID,
			"Sorry, there was an error updating your settings. Please try again.")
		return
	}

	// Show confirmation and return to settings
	status := "disabled"
	emoji := "❌"
	if newState {
		status = "enabled"
		emoji = "✅"
	}

	confirmText := fmt.Sprintf("%s Smart reminders %s! Returning to settings...", emoji, status)
	h.bot.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID, confirmText)

	// Wait briefly, then show settings again
	time.Sleep(1 * time.Second)
	h.handleMenuSettings(ctx, callback, user)
}

// sendQuestionAsEdit sends a learning question by editing an existing message
func (h *BotHandler) sendQuestionAsEdit(chatID int64, messageID int, session *usecases.LearningSession) {
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

	h.bot.EditMessageWithKeyboard(chatID, messageID, fullText, keyboard)
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
