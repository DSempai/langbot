package usecases

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"dutch-learning-bot/internal/domain/grammar"
	"dutch-learning-bot/internal/domain/learning"
	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/domain/vocabulary"
)

// LearningUseCase handles learning-related business operations
type LearningUseCase struct {
	learningRepo    learning.Repository
	vocabularyRepo  vocabulary.Repository
	userRepo        user.Repository
	grammarRepo     grammar.Repository
	preferencesRepo user.PreferencesRepository
}

// NewLearningUseCase creates a new learning use case
func NewLearningUseCase(
	learningRepo learning.Repository,
	vocabularyRepo vocabulary.Repository,
	userRepo user.Repository,
	grammarRepo grammar.Repository,
	preferencesRepo user.PreferencesRepository,
) *LearningUseCase {
	return &LearningUseCase{
		learningRepo:    learningRepo,
		vocabularyRepo:  vocabularyRepo,
		userRepo:        userRepo,
		grammarRepo:     grammarRepo,
		preferencesRepo: preferencesRepo,
	}
}

// LearningSession represents an active learning session
type LearningSession struct {
	UserID       user.ID
	Word         *vocabulary.Word
	Progress     *learning.UserProgress
	QuestionType QuestionType
	StartTime    time.Time
	Options      []string
	CorrectIndex int
	GrammarTip   *grammar.GrammarTip // Optional grammar tip
}

// QuestionType represents the type of question being asked
type QuestionType string

const (
	QuestionTypeEnglishToDutch QuestionType = "english_to_dutch"
	QuestionTypeDutchToEnglish QuestionType = "dutch_to_english"
)

// GetNextDueWord retrieves the next word due for review
func (uc *LearningUseCase) GetNextDueWord(ctx context.Context, userID user.ID) (*LearningSession, error) {
	// Get due words for the user
	progressList, err := uc.learningRepo.FindDueWords(ctx, userID, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get due words: %w", err)
	}

	if len(progressList) == 0 {
		return nil, nil // No words due
	}

	progress := progressList[0]

	// Get the word details
	word, err := uc.vocabularyRepo.FindByID(ctx, progress.WordID())
	if err != nil {
		return nil, fmt.Errorf("failed to get word: %w", err)
	}

	// Randomly choose question type
	questionType := QuestionTypeEnglishToDutch
	if time.Now().UnixNano()%2 == 0 {
		questionType = QuestionTypeDutchToEnglish
	}

	// Generate multiple choice options
	options, correctIndex, err := uc.generateMultipleChoiceOptions(ctx, word, questionType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate options: %w", err)
	}

	session := &LearningSession{
		UserID:       userID,
		Word:         word,
		Progress:     progress,
		QuestionType: questionType,
		StartTime:    time.Now(),
		Options:      options,
		CorrectIndex: correctIndex,
	}

	// Check if user has grammar tips enabled before showing them
	preferences, err := uc.preferencesRepo.FindPreferences(ctx, userID)
	if err == nil && preferences != nil && preferences.GrammarTipsEnabled() {
		// 20% chance to include a contextual grammar tip
		if shouldShowGrammarTip() {
			grammarTip, err := uc.GetContextualGrammarTip(ctx, word, userID)
			if err == nil && grammarTip != nil {
				session.GrammarTip = grammarTip
			}
			// If there's an error getting grammar tip, we just continue without it
		}
	}

	return session, nil
}

// GetContextualGrammarTip gets a grammar tip that's relevant to the current word
func (uc *LearningUseCase) GetContextualGrammarTip(ctx context.Context, word *vocabulary.Word, userID user.ID) (*grammar.GrammarTip, error) {
	// First try to find tips that specifically apply to this word
	applicableTips, err := uc.grammarRepo.FindApplicableToWord(ctx, word.Dutch(), word.English(), string(word.Category()))
	if err != nil {
		return nil, fmt.Errorf("failed to find applicable grammar tips: %w", err)
	}

	if len(applicableTips) > 0 {
		// Return a random applicable tip (no difficulty filtering)
		randomIndex := time.Now().UnixNano() % int64(len(applicableTips))
		return applicableTips[randomIndex], nil
	}

	// If no applicable tips found, don't show a tip (better than irrelevant tip)
	return nil, nil
}

// shouldShowGrammarTip determines if we should show a grammar tip (20% chance)
func shouldShowGrammarTip() bool {
	randomNum, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		// Fallback to time-based randomization
		return time.Now().UnixNano()%100 < 20
	}
	return randomNum.Int64() < 20
}

// generateMultipleChoiceOptions generates 4 options with one correct answer
func (uc *LearningUseCase) generateMultipleChoiceOptions(ctx context.Context, word *vocabulary.Word, questionType QuestionType) ([]string, int, error) {
	// Get all words from the same category for wrong options
	categoryWords, err := uc.vocabularyRepo.FindByCategory(ctx, word.Category())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get category words: %w", err)
	}

	var correctAnswer string
	var wrongAnswers []string

	if questionType == QuestionTypeEnglishToDutch {
		correctAnswer = word.Dutch()
		for _, w := range categoryWords {
			if w.ID() != word.ID() && w.Dutch() != correctAnswer {
				wrongAnswers = append(wrongAnswers, w.Dutch())
			}
		}
	} else {
		correctAnswer = word.English()
		for _, w := range categoryWords {
			if w.ID() != word.ID() && w.English() != correctAnswer {
				wrongAnswers = append(wrongAnswers, w.English())
			}
		}
	}

	// If we don't have enough words in the category, get from all words
	if len(wrongAnswers) < 3 {
		allWords, err := uc.vocabularyRepo.FindAll(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get all words: %w", err)
		}

		for _, w := range allWords {
			if w.ID() != word.ID() {
				var candidate string
				if questionType == QuestionTypeEnglishToDutch {
					candidate = w.Dutch()
				} else {
					candidate = w.English()
				}

				if candidate != correctAnswer {
					// Check if we already have this answer
					found := false
					for _, existing := range wrongAnswers {
						if existing == candidate {
							found = true
							break
						}
					}
					if !found {
						wrongAnswers = append(wrongAnswers, candidate)
					}
				}
			}
			if len(wrongAnswers) >= 3 {
				break
			}
		}
	}

	// Ensure we have at least 3 wrong answers
	if len(wrongAnswers) < 3 {
		return nil, 0, fmt.Errorf("not enough words to generate options")
	}

	// Select 3 random wrong answers
	selectedWrong := make([]string, 3)
	if len(wrongAnswers) >= 3 {
		// Better shuffling using crypto/rand
		for i := len(wrongAnswers) - 1; i > 0; i-- {
			j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
			if err != nil {
				// Fallback to time-based if crypto/rand fails
				j = big.NewInt(int64(time.Now().UnixNano() % int64(i+1)))
			}
			wrongAnswers[i], wrongAnswers[j.Int64()] = wrongAnswers[j.Int64()], wrongAnswers[i]
		}
		copy(selectedWrong, wrongAnswers[:3])
	}

	// Create options array with correct answer at random position
	options := make([]string, 4)
	correctIndexBig, err := rand.Int(rand.Reader, big.NewInt(4))
	correctIndex := int(correctIndexBig.Int64())
	if err != nil {
		// Fallback to time-based if crypto/rand fails
		correctIndex = int(time.Now().UnixNano()) % 4
	}

	options[correctIndex] = correctAnswer
	wrongIndex := 0
	for i := 0; i < 4; i++ {
		if i != correctIndex {
			options[i] = selectedWrong[wrongIndex]
			wrongIndex++
		}
	}

	return options, correctIndex, nil
}

// CheckMultipleChoiceAnswer checks if the selected option index is correct
func (uc *LearningUseCase) CheckMultipleChoiceAnswer(session *LearningSession, selectedIndex int) bool {
	return selectedIndex == session.CorrectIndex
}

// ProcessReview processes a user's review of a word
func (uc *LearningUseCase) ProcessReview(
	ctx context.Context,
	session *LearningSession,
	rating learning.Rating,
	responseTime time.Duration,
) error {
	// Process the review
	result := session.Progress.Review(rating)

	// Save or update progress in repository
	var err error
	if session.Progress.ID() == 0 {
		// This is a new word, save it to the database
		err = uc.learningRepo.SaveProgress(ctx, session.Progress)
		if err != nil {
			return fmt.Errorf("failed to save new progress: %w", err)
		}
	} else {
		// This is an existing word, update it
		err = uc.learningRepo.UpdateProgress(ctx, session.Progress)
		if err != nil {
			return fmt.Errorf("failed to update progress: %w", err)
		}
	}

	// Save review history
	history := learning.NewReviewHistory(
		session.UserID,
		session.Word.ID(),
		rating,
		responseTime,
	)

	err = uc.learningRepo.SaveReviewHistory(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to save review history: %w", err)
	}

	// Update result with the new card state (if needed for further processing)
	_ = result

	return nil
}

// GetOrCreateProgress gets existing progress or creates new progress for a user-word pair
func (uc *LearningUseCase) GetOrCreateProgress(
	ctx context.Context,
	userID user.ID,
	wordID vocabulary.ID,
) (*learning.UserProgress, error) {
	// Try to find existing progress
	progress, err := uc.learningRepo.FindProgress(ctx, userID, wordID)
	if err != nil {
		return nil, fmt.Errorf("failed to find progress: %w", err)
	}

	// If no progress exists, create new one
	if progress == nil {
		progress = learning.NewUserProgress(userID, wordID)
		err = uc.learningRepo.SaveProgress(ctx, progress)
		if err != nil {
			return nil, fmt.Errorf("failed to save new progress: %w", err)
		}
	}

	return progress, nil
}

// GetUserStats retrieves learning statistics for a user
func (uc *LearningUseCase) GetUserStats(ctx context.Context, userID user.ID) (*learning.UserStats, error) {
	stats, err := uc.learningRepo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}

// CheckAnswer checks if the user's answer is correct
func (uc *LearningUseCase) CheckAnswer(session *LearningSession, userAnswer string) bool {
	var correctAnswer string

	switch session.QuestionType {
	case QuestionTypeEnglishToDutch:
		correctAnswer = session.Word.Dutch()
	case QuestionTypeDutchToEnglish:
		correctAnswer = session.Word.English()
	}

	// Simple case-insensitive comparison
	// Could be enhanced with fuzzy matching, accent handling, etc.
	return normalizeAnswer(userAnswer) == normalizeAnswer(correctAnswer)
}

// normalizeAnswer normalizes an answer for comparison
func normalizeAnswer(answer string) string {
	// Convert to lowercase and trim whitespace
	// Could be enhanced with more sophisticated normalization
	return strings.ToLower(strings.TrimSpace(answer))
}
