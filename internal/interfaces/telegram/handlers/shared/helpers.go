package shared

import (
	"fmt"
	"strings"

	"dutch-learning-bot/internal/domain/learning"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CreateMainMenuKeyboard creates the standard main menu keyboard
func CreateMainMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Start Learning", "menu_learn"),
			tgbotapi.NewInlineKeyboardButtonData("📊 View Stats", "menu_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "menu_help"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "menu_settings"),
		),
	)
}

// CreateStatsKeyboard creates a keyboard for stats view
func CreateStatsKeyboard(isCallback bool) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📚 Start Learning", "menu_learn"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
		),
	)
}

// CreateHelpKeyboard creates a keyboard for help view
func CreateHelpKeyboard(isCallback bool) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
		),
	)
}

// CreateNoWordsKeyboard creates a keyboard for when no words are available
func CreateNoWordsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 View Stats", "menu_stats"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "back_menu"),
		),
	)
}

// FormatStatsText formats user statistics into a readable message
func FormatStatsText(stats *learning.UserStats) string {
	return fmt.Sprintf(
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
}

// GetHelpText returns the standard help text
func GetHelpText() string {
	return `🇳🇱 **Dutch Learning Bot Help**

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
- Use the Settings menu to customize your learning experience

Good luck with your Dutch learning! 🍀`
}

// EscapeMarkdown escapes special Markdown characters
func EscapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}
