package learning

import (
	"context"

	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/domain/vocabulary"
)

// Repository defines the contract for learning progress persistence
type Repository interface {
	// SaveProgress persists user progress
	SaveProgress(ctx context.Context, progress *UserProgress) error

	// UpdateProgress updates existing user progress
	UpdateProgress(ctx context.Context, progress *UserProgress) error

	// FindProgress retrieves user progress for a specific word
	FindProgress(ctx context.Context, userID user.ID, wordID vocabulary.ID) (*UserProgress, error)

	// FindDueWords retrieves words that are due for review for a user
	FindDueWords(ctx context.Context, userID user.ID, limit int) ([]*UserProgress, error)

	// FindNewWords retrieves words that don't have progress records yet
	FindNewWords(ctx context.Context, userID user.ID, limit int) ([]*UserProgress, error)

	// FindProgressByUser retrieves all progress for a user
	FindProgressByUser(ctx context.Context, userID user.ID) ([]*UserProgress, error)

	// SaveReviewHistory persists review history
	SaveReviewHistory(ctx context.Context, history *ReviewHistory) error

	// FindReviewHistory retrieves review history for a user and word
	FindReviewHistory(ctx context.Context, userID user.ID, wordID vocabulary.ID) ([]*ReviewHistory, error)

	// GetUserStats retrieves learning statistics for a user
	GetUserStats(ctx context.Context, userID user.ID) (*UserStats, error)

	// GetUsersWithProgress retrieves all users who have learning progress
	GetUsersWithProgress(ctx context.Context) ([]user.ID, error)

	// SaveProgressAndHistory persists both user progress and review history
	SaveProgressAndHistory(ctx context.Context, progress *UserProgress, history *ReviewHistory) error
}

// UserStats represents learning statistics for a user
type UserStats struct {
	TotalWords     int
	NewWords       int
	LearningWords  int
	ReviewWords    int
	DueWords       int
	AvgDifficulty  float64
	TotalReviews   int
	CorrectReviews int
}
