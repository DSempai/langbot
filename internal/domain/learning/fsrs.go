package learning

import (
	"math"
	"time"
)

// FSRS parameters - these can be tuned based on user performance
const (
	// Default parameters for FSRS v4
	defaultWeight0  = 0.4072
	defaultWeight1  = 1.1829
	defaultWeight2  = 3.1262
	defaultWeight3  = 15.4722
	defaultWeight4  = 7.2102
	defaultWeight5  = 0.5316
	defaultWeight6  = 1.0651
	defaultWeight7  = 0.0234
	defaultWeight8  = 1.616
	defaultWeight9  = 0.1544
	defaultWeight10 = 1.0824
	defaultWeight11 = 1.9813
	defaultWeight12 = 0.0953
	defaultWeight13 = 0.2975
	defaultWeight14 = 2.2042
	defaultWeight15 = 0.2407
	defaultWeight16 = 2.9466
	defaultWeight17 = 0.5034
	defaultWeight18 = 0.6567

	// Decay parameter for memory strength
	decayParam = -0.5
	// Factor for calculating next review interval
	factor = 19.0 / 81.0
	// Request retention (target recall probability)
	requestRetention = 0.9
)

// FSRSCard represents the state of a card in FSRS
type FSRSCard struct {
	dueDate     time.Time
	stability   float64
	difficulty  float64
	lastReview  time.Time
	state       State
	reviewCount int
	lapses      int
}

// State represents the learning state of a card
type State string

const (
	StateNew        State = "new"
	StateLearning   State = "learning"
	StateReview     State = "review"
	StateRelearning State = "relearning"
)

// Rating represents user's performance rating
type Rating int

const (
	Again Rating = 1 // Complete blackout
	Hard  Rating = 2 // Incorrect response; the correct one remembered upon seeing the answer
	Good  Rating = 3 // Correct response after a hesitation
	Easy  Rating = 4 // Perfect response
)

// ReviewResult contains the updated card state after review
type ReviewResult struct {
	Card     *FSRSCard
	LogEntry *ReviewLog
}

// ReviewLog represents a single review entry
type ReviewLog struct {
	Rating        Rating
	ScheduledDays int
	ElapsedDays   int
	ReviewTime    time.Time
	State         State
}

// NewFSRSCard creates a new card with default FSRS parameters
func NewFSRSCard() *FSRSCard {
	return &FSRSCard{
		dueDate:     time.Now(),
		stability:   1.0,
		difficulty:  5.0,
		state:       StateNew,
		reviewCount: 0,
		lapses:      0,
	}
}

// Getters
func (card *FSRSCard) DueDate() time.Time    { return card.dueDate }
func (card *FSRSCard) Stability() float64    { return card.stability }
func (card *FSRSCard) Difficulty() float64   { return card.difficulty }
func (card *FSRSCard) LastReview() time.Time { return card.lastReview }
func (card *FSRSCard) State() State          { return card.state }
func (card *FSRSCard) ReviewCount() int      { return card.reviewCount }
func (card *FSRSCard) Lapses() int           { return card.lapses }

// IsDue checks if the card is due for review
func (card *FSRSCard) IsDue() bool {
	return time.Now().After(card.dueDate) || time.Now().Equal(card.dueDate)
}

// Review processes a review and returns updated card state
func (card *FSRSCard) Review(rating Rating, reviewTime time.Time) *ReviewResult {
	elapsed := int(reviewTime.Sub(card.lastReview).Hours() / 24)
	if elapsed < 0 {
		elapsed = 0
	}

	var scheduled int
	if !card.dueDate.IsZero() {
		scheduled = int(card.dueDate.Sub(card.lastReview).Hours() / 24)
		if scheduled < 0 {
			scheduled = 0
		}
	}

	// Create a copy of the card with updated review info
	newCard := *card
	newCard.lastReview = reviewTime
	newCard.reviewCount++

	// Apply state-specific review logic
	switch card.state {
	case StateNew:
		stateCard := card.reviewNew(rating)
		// Preserve the updated review count and last review time
		stateCard.lastReview = reviewTime
		stateCard.reviewCount = card.reviewCount + 1
		newCard = stateCard
	case StateLearning, StateRelearning:
		stateCard := card.reviewLearning(rating)
		// Preserve the updated review count and last review time
		stateCard.lastReview = reviewTime
		stateCard.reviewCount = card.reviewCount + 1
		newCard = stateCard
	case StateReview:
		stateCard := card.reviewReview(rating, elapsed)
		// Preserve the updated review count and last review time
		stateCard.lastReview = reviewTime
		stateCard.reviewCount = card.reviewCount + 1
		newCard = stateCard
	}

	logEntry := &ReviewLog{
		Rating:        rating,
		ScheduledDays: scheduled,
		ElapsedDays:   elapsed,
		ReviewTime:    reviewTime,
		State:         card.state,
	}

	return &ReviewResult{
		Card:     &newCard,
		LogEntry: logEntry,
	}
}

func (card *FSRSCard) reviewNew(rating Rating) FSRSCard {
	newCard := *card
	newCard.difficulty = initDifficulty(rating)

	switch rating {
	case Again:
		newCard.state = StateLearning
		newCard.dueDate = time.Now().Add(1 * time.Minute)
	case Hard:
		newCard.state = StateLearning
		newCard.dueDate = time.Now().Add(5 * time.Minute)
	case Good:
		newCard.state = StateLearning
		newCard.dueDate = time.Now().Add(10 * time.Minute)
	case Easy:
		newCard.state = StateReview
		newCard.stability = initStability(rating)
		interval := calculateInterval(newCard.stability)
		newCard.dueDate = time.Now().Add(time.Duration(interval) * 24 * time.Hour)
	}

	return newCard
}

func (card *FSRSCard) reviewLearning(rating Rating) FSRSCard {
	newCard := *card

	switch rating {
	case Again:
		newCard.state = StateLearning
		newCard.dueDate = time.Now().Add(1 * time.Minute)
	case Hard:
		newCard.state = StateLearning
		newCard.dueDate = time.Now().Add(5 * time.Minute)
	case Good:
		newCard.state = StateReview
		newCard.stability = initStability(Good)
		interval := calculateInterval(newCard.stability)
		newCard.dueDate = time.Now().Add(time.Duration(interval) * 24 * time.Hour)
	case Easy:
		newCard.state = StateReview
		newCard.stability = initStability(Easy)
		interval := calculateInterval(newCard.stability)
		newCard.dueDate = time.Now().Add(time.Duration(interval) * 24 * time.Hour)
	}

	return newCard
}

func (card *FSRSCard) reviewReview(rating Rating, elapsed int) FSRSCard {
	newCard := *card

	if rating == Again {
		newCard.lapses++
		newCard.state = StateRelearning
		newCard.dueDate = time.Now().Add(5 * time.Minute)
	} else {
		newCard.state = StateReview
		newCard.stability = nextStability(card.difficulty, card.stability, rating)
		newCard.difficulty = nextDifficulty(card.difficulty, rating)
		interval := calculateInterval(newCard.stability)
		newCard.dueDate = time.Now().Add(time.Duration(interval) * 24 * time.Hour)
	}

	return newCard
}

// initDifficulty calculates initial difficulty based on rating
func initDifficulty(rating Rating) float64 {
	return math.Max(defaultWeight4-defaultWeight5*float64(rating-3), 1.0)
}

// initStability calculates initial stability based on rating
func initStability(rating Rating) float64 {
	return math.Max(defaultWeight0+defaultWeight1*float64(rating-1), 0.1)
}

// nextStability calculates next stability value
func nextStability(difficulty, stability float64, rating Rating) float64 {
	hardPenalty := 1.0
	if rating == Hard {
		hardPenalty = defaultWeight6
	}

	easyBonus := 1.0
	if rating == Easy {
		easyBonus = defaultWeight7
	}

	return stability * (1 + math.Exp(defaultWeight8)*
		(11-difficulty)*
		math.Pow(stability, defaultWeight9)*
		(math.Exp((1-requestRetention)*defaultWeight10)-1)*
		hardPenalty*
		easyBonus)
}

// nextDifficulty calculates next difficulty value
func nextDifficulty(difficulty float64, rating Rating) float64 {
	deltaD := -defaultWeight11 * (float64(rating) - 3)
	newDifficulty := difficulty + deltaD

	// Mean reversion to 5.0
	meanReversion := defaultWeight12 * (5.0 - newDifficulty)
	newDifficulty += meanReversion

	return math.Max(math.Min(newDifficulty, 10.0), 1.0)
}

// calculateInterval calculates review interval based on stability
func calculateInterval(stability float64) int {
	interval := stability * math.Log(requestRetention) / math.Log(0.9)
	return int(math.Max(math.Round(interval), 1))
}

// Setters for restoring from database
func (card *FSRSCard) SetDueDate(dueDate time.Time)       { card.dueDate = dueDate }
func (card *FSRSCard) SetStability(stability float64)     { card.stability = stability }
func (card *FSRSCard) SetDifficulty(difficulty float64)   { card.difficulty = difficulty }
func (card *FSRSCard) SetLastReview(lastReview time.Time) { card.lastReview = lastReview }
func (card *FSRSCard) SetState(state State)               { card.state = state }
func (card *FSRSCard) SetReviewCount(count int)           { card.reviewCount = count }
func (card *FSRSCard) SetLapses(lapses int)               { card.lapses = lapses }
