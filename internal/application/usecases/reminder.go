package usecases

import (
	"context"
	"fmt"
	"log"
	"time"

	"dutch-learning-bot/internal/domain/learning"
	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/infrastructure/telegram"
)

// ReminderConfig holds configuration for the reminder system
type ReminderConfig struct {
	// How often to check for reminders
	CheckInterval time.Duration
	// Minimum time between reminders for the same user
	MinReminderInterval time.Duration
	// Hours of day when reminders are sent (24-hour format)
	QuietHoursStart int
	QuietHoursEnd   int
	// Maximum reminders per day per user
	MaxRemindersPerDay int
}

// DefaultReminderConfig returns sensible defaults for reminders
func DefaultReminderConfig() *ReminderConfig {
	return &ReminderConfig{
		CheckInterval:       30 * time.Minute, // Check every 30 minutes
		MinReminderInterval: 4 * time.Hour,    // Don't remind more than once every 4 hours
		QuietHoursStart:     22,               // 10 PM
		QuietHoursEnd:       8,                // 8 AM
		MaxRemindersPerDay:  3,                // Max 3 reminders per day
	}
}

// ReminderUseCase handles smart reminder functionality
type ReminderUseCase struct {
	bot           *telegram.Bot
	userRepo      user.Repository
	learningRepo  learning.Repository
	config        *ReminderConfig
	reminderState map[user.ID]*UserReminderState
}

// UserReminderState tracks reminder state for each user
type UserReminderState struct {
	LastReminderSent time.Time
	RemindersToday   int
	LastCheckDate    time.Time
}

// NewReminderUseCase creates a new reminder use case
func NewReminderUseCase(
	bot *telegram.Bot,
	userRepo user.Repository,
	learningRepo learning.Repository,
	config *ReminderConfig,
) *ReminderUseCase {
	if config == nil {
		config = DefaultReminderConfig()
	}

	return &ReminderUseCase{
		bot:           bot,
		userRepo:      userRepo,
		learningRepo:  learningRepo,
		config:        config,
		reminderState: make(map[user.ID]*UserReminderState),
	}
}

// StartReminderService begins the background reminder service
func (uc *ReminderUseCase) StartReminderService(ctx context.Context) {
	log.Printf("Starting smart reminder service (check interval: %v)", uc.config.CheckInterval)

	ticker := time.NewTicker(uc.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Reminder service stopping...")
			return
		case <-ticker.C:
			uc.checkAndSendReminders(ctx)
		}
	}
}

// checkAndSendReminders checks for users needing reminders and sends them
func (uc *ReminderUseCase) checkAndSendReminders(ctx context.Context) {
	log.Printf("Checking for users needing reminders...")

	// Get all users who have used the bot (have progress records)
	users, err := uc.getUsersWithProgress(ctx)
	if err != nil {
		log.Printf("Failed to get users with progress: %v", err)
		return
	}

	remindersSent := 0
	for _, u := range users {
		if uc.shouldSendReminder(ctx, u) {
			if uc.sendReminderToUser(ctx, u) {
				remindersSent++
			}
		}
	}

	if remindersSent > 0 {
		log.Printf("Sent %d smart reminders", remindersSent)
	}
}

// shouldSendReminder determines if a user should receive a reminder
func (uc *ReminderUseCase) shouldSendReminder(ctx context.Context, u *user.User) bool {
	now := time.Now()
	userID := u.ID()

	// Check quiet hours
	if uc.isQuietTime(now) {
		return false
	}

	// Get or create reminder state for this user
	state, exists := uc.reminderState[userID]
	if !exists {
		state = &UserReminderState{
			LastCheckDate: now.AddDate(0, 0, -1), // Set to yesterday to reset counter
		}
		uc.reminderState[userID] = state
	}

	// Reset daily counter if it's a new day
	if !isSameDay(state.LastCheckDate, now) {
		state.RemindersToday = 0
		state.LastCheckDate = now
	}

	// Check if we've exceeded daily limit
	if state.RemindersToday >= uc.config.MaxRemindersPerDay {
		return false
	}

	// Check minimum interval between reminders
	if now.Sub(state.LastReminderSent) < uc.config.MinReminderInterval {
		return false
	}

	// Check if user has due words
	stats, err := uc.learningRepo.GetUserStats(ctx, userID)
	if err != nil {
		log.Printf("Failed to get stats for user %d: %v", userID, err)
		return false
	}

	// Only remind if there are actually due words
	if stats.DueWords == 0 {
		return false
	}

	// Smart logic: Consider user's activity pattern
	// Don't remind users who were recently active (within last hour)
	if now.Sub(u.LastActive()) < time.Hour {
		return false
	}

	// More likely to remind if user hasn't been active for a while
	daysSinceActive := int(now.Sub(u.LastActive()).Hours() / 24)

	// Always remind if user hasn't been active for 3+ days and has due words
	if daysSinceActive >= 3 {
		return true
	}

	// For users active within last 3 days, use a more sophisticated check
	// Consider the number of due words and time since last reminder
	hoursSinceLastReminder := now.Sub(state.LastReminderSent).Hours()

	// If user has many due words (5+), remind sooner
	if stats.DueWords >= 5 && hoursSinceLastReminder >= 6 {
		return true
	}

	// If user has some due words (1-4), remind after longer interval
	if stats.DueWords >= 1 && hoursSinceLastReminder >= 12 {
		return true
	}

	return false
}

// sendReminderToUser sends a smart reminder to a specific user
func (uc *ReminderUseCase) sendReminderToUser(ctx context.Context, u *user.User) bool {
	userID := u.ID()

	// Get current stats
	stats, err := uc.learningRepo.GetUserStats(ctx, userID)
	if err != nil {
		log.Printf("Failed to get stats for user %d: %v", userID, err)
		return false
	}

	// Create personalized reminder message
	reminderText := uc.createReminderMessage(u, stats)

	// Send the reminder
	telegramID := int64(u.TelegramID())
	err = uc.bot.SendMessageWithMarkdown(telegramID, reminderText)
	if err != nil {
		log.Printf("Failed to send reminder to user %d (telegram: %d): %v", userID, telegramID, err)
		return false
	}

	// Update reminder state
	state := uc.reminderState[userID]
	state.LastReminderSent = time.Now()
	state.RemindersToday++

	log.Printf("Sent smart reminder to user %d (%s) - %d due words", userID, u.FirstName(), stats.DueWords)
	return true
}

// createReminderMessage creates a personalized reminder message
func (uc *ReminderUseCase) createReminderMessage(u *user.User, stats *learning.UserStats) string {
	firstName := u.FirstName()
	if firstName == "" {
		firstName = "there"
	}

	// Determine time of day greeting
	hour := time.Now().Hour()
	var greeting string
	switch {
	case hour < 12:
		greeting = "Good morning"
	case hour < 17:
		greeting = "Good afternoon"
	default:
		greeting = "Good evening"
	}

	// Create personalized message based on due words count
	var message string
	switch {
	case stats.DueWords == 1:
		message = fmt.Sprintf(
			"🇳🇱 %s, %s!\n\n"+
				"You have **1 Dutch word** ready for review. "+
				"A quick review now will help strengthen your memory! 🧠\n\n"+
				"Use /learn to practice, or /menu for options.",
			greeting, firstName)

	case stats.DueWords <= 5:
		message = fmt.Sprintf(
			"🇳🇱 %s, %s!\n\n"+
				"You have **%d Dutch words** waiting for review. "+
				"Perfect time for a quick practice session! ✨\n\n"+
				"Use /learn to start, or /menu for more options.",
			greeting, firstName, stats.DueWords)

	case stats.DueWords <= 10:
		message = fmt.Sprintf(
			"🇳🇱 %s, %s!\n\n"+
				"Great progress! You have **%d words** due for review. "+
				"Reviewing them now will boost your retention significantly! 🚀\n\n"+
				"Use /learn to begin, or /stats to see your progress.",
			greeting, firstName, stats.DueWords)

	default:
		message = fmt.Sprintf(
			"🇳🇱 %s, %s!\n\n"+
				"Wow! You have **%d Dutch words** ready for review. "+
				"This is a great opportunity to reinforce your learning! 💪\n\n"+
				"Don't worry - start with /learn and go at your own pace. Every word counts!",
			greeting, firstName, stats.DueWords)
	}

	// Add motivational elements based on progress
	if stats.ReviewWords > 0 {
		message += fmt.Sprintf("\n\n📊 You've mastered **%d words** so far - keep it up! 🌟", stats.ReviewWords)
	}

	return message
}

// getUsersWithProgress gets all users who have made progress (have used the bot)
func (uc *ReminderUseCase) getUsersWithProgress(ctx context.Context) ([]*user.User, error) {
	// This is a simplified approach - in a real implementation, you might want
	// to add a method to get active users directly from the repository
	// For now, we'll get users from the learning repository who have progress
	return uc.getAllUsersWithLearningProgress(ctx)
}

// getAllUsersWithLearningProgress gets users who have learning progress
func (uc *ReminderUseCase) getAllUsersWithLearningProgress(ctx context.Context) ([]*user.User, error) {
	// Get user IDs who have learning progress
	userIDs, err := uc.learningRepo.GetUsersWithProgress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users with progress: %w", err)
	}

	// Get full user objects
	var users []*user.User
	for _, userID := range userIDs {
		u, err := uc.userRepo.FindByID(ctx, userID)
		if err != nil {
			log.Printf("Failed to get user %d: %v", userID, err)
			continue
		}
		if u != nil {
			users = append(users, u)
		}
	}

	return users, nil
}

// isQuietTime checks if current time is within quiet hours
func (uc *ReminderUseCase) isQuietTime(t time.Time) bool {
	hour := t.Hour()
	start := uc.config.QuietHoursStart
	end := uc.config.QuietHoursEnd

	if start <= end {
		// Normal case: e.g., 22:00 to 08:00 next day
		return hour >= start || hour < end
	} else {
		// Quiet hours cross midnight: e.g., 10:00 to 06:00
		return hour >= start && hour < end
	}
}

// isSameDay checks if two times are on the same day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// GetReminderStats returns statistics about reminders for debugging
func (uc *ReminderUseCase) GetReminderStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["total_users_tracked"] = len(uc.reminderState)
	stats["config"] = uc.config

	todayReminders := 0
	for _, state := range uc.reminderState {
		if isSameDay(state.LastCheckDate, time.Now()) {
			todayReminders += state.RemindersToday
		}
	}
	stats["reminders_sent_today"] = todayReminders

	return stats
}
