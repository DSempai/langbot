package learning

import (
	"time"

	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/domain/vocabulary"
)

// UserProgress represents a user's progress on a specific word
type UserProgress struct {
	id        ID
	userID    user.ID
	wordID    vocabulary.ID
	fsrsCard  *FSRSCard
	createdAt time.Time
	updatedAt time.Time
}

// ID represents the user progress unique identifier
type ID int64

// NewUserProgress creates a new user progress record
func NewUserProgress(userID user.ID, wordID vocabulary.ID) *UserProgress {
	now := time.Now()
	return &UserProgress{
		userID:    userID,
		wordID:    wordID,
		fsrsCard:  NewFSRSCard(),
		createdAt: now,
		updatedAt: now,
	}
}

// Getters
func (up *UserProgress) ID() ID                { return up.id }
func (up *UserProgress) UserID() user.ID       { return up.userID }
func (up *UserProgress) WordID() vocabulary.ID { return up.wordID }
func (up *UserProgress) FSRSCard() *FSRSCard   { return up.fsrsCard }
func (up *UserProgress) CreatedAt() time.Time  { return up.createdAt }
func (up *UserProgress) UpdatedAt() time.Time  { return up.updatedAt }

// SetID sets the progress ID (used by repository)
func (up *UserProgress) SetID(id ID) {
	up.id = id
}

// Review processes a review and updates the FSRS card
func (up *UserProgress) Review(rating Rating) *ReviewResult {
	result := up.fsrsCard.Review(rating, time.Now())
	// Replace the current card with the updated one from the result
	up.fsrsCard = result.Card
	up.updatedAt = time.Now()
	return result
}

// IsDue checks if this word is due for review
func (up *UserProgress) IsDue() bool {
	return up.fsrsCard.IsDue()
}

// ReviewHistory represents a single review attempt
type ReviewHistory struct {
	id             ID
	userID         user.ID
	wordID         vocabulary.ID
	rating         Rating
	reviewTime     time.Time
	responseTimeMs int
}

// NewReviewHistory creates a new review history entry
func NewReviewHistory(userID user.ID, wordID vocabulary.ID, rating Rating, responseTime time.Duration) *ReviewHistory {
	return &ReviewHistory{
		userID:         userID,
		wordID:         wordID,
		rating:         rating,
		reviewTime:     time.Now(),
		responseTimeMs: int(responseTime.Milliseconds()),
	}
}

// Getters for ReviewHistory
func (rh *ReviewHistory) ID() ID                { return rh.id }
func (rh *ReviewHistory) UserID() user.ID       { return rh.userID }
func (rh *ReviewHistory) WordID() vocabulary.ID { return rh.wordID }
func (rh *ReviewHistory) Rating() Rating        { return rh.rating }
func (rh *ReviewHistory) ReviewTime() time.Time { return rh.reviewTime }
func (rh *ReviewHistory) ResponseTimeMs() int   { return rh.responseTimeMs }

// SetID sets the review history ID (used by repository)
func (rh *ReviewHistory) SetID(id ID) {
	rh.id = id
}

// SetReviewTime sets the review time (used by repository when loading from database)
func (rh *ReviewHistory) SetReviewTime(reviewTime time.Time) {
	rh.reviewTime = reviewTime
}
