package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"dutch-learning-bot/internal/domain/learning"
	"dutch-learning-bot/internal/domain/user"
	"dutch-learning-bot/internal/domain/vocabulary"
)

type learningRepository struct {
	db *sql.DB
}

// NewLearningRepository creates a new learning repository
func NewLearningRepository(db *sql.DB) learning.Repository {
	return &learningRepository{db: db}
}

// SaveProgress persists user progress
func (r *learningRepository) SaveProgress(ctx context.Context, progress *learning.UserProgress) error {
	query := `
		INSERT INTO user_progress 
		(user_id, word_id, stability, difficulty, last_review, due_date, review_count, lapses, state, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	fsrsCard := progress.FSRSCard()
	result, err := r.db.ExecContext(ctx, query,
		int64(progress.UserID()), int64(progress.WordID()),
		fsrsCard.Stability(), fsrsCard.Difficulty(),
		fsrsCard.LastReview(), fsrsCard.DueDate(),
		fsrsCard.ReviewCount(), fsrsCard.Lapses(), string(fsrsCard.State()),
		progress.CreatedAt(), progress.UpdatedAt())

	if err != nil {
		return fmt.Errorf("failed to save progress: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get progress ID: %w", err)
	}

	progress.SetID(learning.ID(id))
	return nil
}

// UpdateProgress updates existing user progress
func (r *learningRepository) UpdateProgress(ctx context.Context, progress *learning.UserProgress) error {
	query := `
		UPDATE user_progress 
		SET stability = ?, difficulty = ?, last_review = ?, due_date = ?, 
		    review_count = ?, lapses = ?, state = ?, updated_at = ?
		WHERE id = ?
	`

	fsrsCard := progress.FSRSCard()
	_, err := r.db.ExecContext(ctx, query,
		fsrsCard.Stability(), fsrsCard.Difficulty(),
		fsrsCard.LastReview(), fsrsCard.DueDate(),
		fsrsCard.ReviewCount(), fsrsCard.Lapses(), string(fsrsCard.State()),
		progress.UpdatedAt(), int64(progress.ID()))

	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	return nil
}

// FindProgress retrieves user progress for a specific word
func (r *learningRepository) FindProgress(ctx context.Context, userID user.ID, wordID vocabulary.ID) (*learning.UserProgress, error) {
	query := `
		SELECT id, user_id, word_id, stability, difficulty, last_review, due_date, 
		       review_count, lapses, state, created_at, updated_at
		FROM user_progress 
		WHERE user_id = ? AND word_id = ?
	`

	var id learning.ID
	var uID user.ID
	var wID vocabulary.ID
	var stability, difficulty float64
	var lastReviewStr, dueDateStr, createdAtStr, updatedAtStr sql.NullString
	var reviewCount, lapses int
	var state string

	err := r.db.QueryRowContext(ctx, query, int64(userID), int64(wordID)).Scan(
		&id, &uID, &wID, &stability, &difficulty, &lastReviewStr, &dueDateStr,
		&reviewCount, &lapses, &state, &createdAtStr, &updatedAtStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find progress: %w", err)
	}

	progress := learning.NewUserProgress(userID, wordID)
	progress.SetID(id)

	// Parse datetime strings
	lastReview, err := r.parseDateTime(lastReviewStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_review: %w", err)
	}

	dueDate, err := r.parseDateTime(dueDateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse due_date: %w", err)
	}

	// Parse but don't use createdAt and updatedAt since they're not used in this context
	_, err = r.parseDateTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	_, err = r.parseDateTime(updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	// Reconstruct FSRS card from database data
	fsrsCard := progress.FSRSCard()
	r.setFSRSCardFromDB(fsrsCard, stability, difficulty, lastReview, dueDate, reviewCount, lapses, state)

	return progress, nil
}

// FindDueWords retrieves words that are due for review for a user
func (r *learningRepository) FindDueWords(ctx context.Context, userID user.ID, limit int) ([]*learning.UserProgress, error) {
	query := `
		SELECT id, user_id, word_id, stability, difficulty, last_review, due_date, 
		       review_count, lapses, state, created_at, updated_at
		FROM user_progress 
		WHERE user_id = ? AND due_date <= CURRENT_TIMESTAMP
		ORDER BY due_date ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, int64(userID), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query due progress words: %w", err)
	}
	defer rows.Close()

	var progressList []*learning.UserProgress
	for rows.Next() {
		progress, err := r.scanProgressRow(rows, userID)
		if err != nil {
			return nil, err
		}
		progressList = append(progressList, progress)
	}

	return progressList, rows.Err()
}

// FindNewWords gets words that don't have progress records yet
func (r *learningRepository) FindNewWords(ctx context.Context, userID user.ID, limit int) ([]*learning.UserProgress, error) {
	query := `
		SELECT w.id as word_id
		FROM words w
		WHERE w.id NOT IN (SELECT word_id FROM user_progress WHERE user_id = ?)
		ORDER BY RANDOM()
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, int64(userID), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query new words: %w", err)
	}
	defer rows.Close()

	var progressList []*learning.UserProgress
	for rows.Next() {
		var wordID vocabulary.ID
		if err := rows.Scan(&wordID); err != nil {
			return nil, fmt.Errorf("failed to scan word ID: %w", err)
		}

		// Create new progress for this word
		progress := learning.NewUserProgress(userID, wordID)
		progressList = append(progressList, progress)
	}

	return progressList, rows.Err()
}

// scanProgressRow scans a progress row from the database
func (r *learningRepository) scanProgressRow(rows *sql.Rows, userID user.ID) (*learning.UserProgress, error) {
	var id learning.ID
	var uID user.ID
	var wID vocabulary.ID
	var stability, difficulty float64
	var lastReviewStr, dueDateStr, createdAtStr, updatedAtStr sql.NullString
	var reviewCount, lapses int
	var state string

	err := rows.Scan(&id, &uID, &wID, &stability, &difficulty, &lastReviewStr, &dueDateStr,
		&reviewCount, &lapses, &state, &createdAtStr, &updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to scan progress: %w", err)
	}

	// Parse datetime strings
	lastReview, err := r.parseDateTime(lastReviewStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_review: %w", err)
	}

	dueDate, err := r.parseDateTime(dueDateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse due_date: %w", err)
	}

	// Parse but don't use createdAt and updatedAt since they're not used in this context
	_, err = r.parseDateTime(createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	_, err = r.parseDateTime(updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %w", err)
	}

	progress := learning.NewUserProgress(userID, wID)
	progress.SetID(id)

	// Set FSRS card data
	fsrsCard := progress.FSRSCard()
	r.setFSRSCardFromDB(fsrsCard, stability, difficulty, lastReview, dueDate, reviewCount, lapses, state)

	return progress, nil
}

// FindProgressByUser retrieves all progress for a user
func (r *learningRepository) FindProgressByUser(ctx context.Context, userID user.ID) ([]*learning.UserProgress, error) {
	query := `
		SELECT id, user_id, word_id, stability, difficulty, last_review, due_date, 
		       review_count, lapses, state, created_at, updated_at
		FROM user_progress 
		WHERE user_id = ?
		ORDER BY updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, int64(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to query user progress: %w", err)
	}
	defer rows.Close()

	var progressList []*learning.UserProgress

	for rows.Next() {
		var id learning.ID
		var uID user.ID
		var wID vocabulary.ID
		var stability, difficulty float64
		var lastReviewStr, dueDateStr, createdAtStr, updatedAtStr sql.NullString
		var reviewCount, lapses int
		var state string

		err := rows.Scan(&id, &uID, &wID, &stability, &difficulty, &lastReviewStr, &dueDateStr,
			&reviewCount, &lapses, &state, &createdAtStr, &updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan progress: %w", err)
		}

		// Parse datetime strings
		lastReview, err := r.parseDateTime(lastReviewStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_review: %w", err)
		}

		dueDate, err := r.parseDateTime(dueDateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse due_date: %w", err)
		}

		// Parse but don't use createdAt and updatedAt since they're not used in this context
		_, err = r.parseDateTime(createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		_, err = r.parseDateTime(updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		progress := learning.NewUserProgress(userID, wID)
		progress.SetID(id)

		// Set FSRS card data
		fsrsCard := progress.FSRSCard()
		r.setFSRSCardFromDB(fsrsCard, stability, difficulty, lastReview, dueDate, reviewCount, lapses, state)

		progressList = append(progressList, progress)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return progressList, nil
}

// SaveReviewHistory persists review history
func (r *learningRepository) SaveReviewHistory(ctx context.Context, history *learning.ReviewHistory) error {
	query := `
		INSERT INTO review_history (user_id, word_id, rating, review_time, response_time_ms)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		int64(history.UserID()), int64(history.WordID()),
		int(history.Rating()), history.ReviewTime(), history.ResponseTimeMs())

	if err != nil {
		return fmt.Errorf("failed to save review history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get review history ID: %w", err)
	}

	history.SetID(learning.ID(id))
	return nil
}

// FindReviewHistory retrieves review history for a user and word
func (r *learningRepository) FindReviewHistory(ctx context.Context, userID user.ID, wordID vocabulary.ID) ([]*learning.ReviewHistory, error) {
	query := `
		SELECT id, user_id, word_id, rating, review_time, response_time_ms
		FROM review_history 
		WHERE user_id = ? AND word_id = ?
		ORDER BY review_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, int64(userID), int64(wordID))
	if err != nil {
		return nil, fmt.Errorf("failed to query review history: %w", err)
	}
	defer rows.Close()

	var historyList []*learning.ReviewHistory

	for rows.Next() {
		var id learning.ID
		var uID user.ID
		var wID vocabulary.ID
		var rating int
		var reviewTimeStr sql.NullString
		var responseTimeMs int

		err := rows.Scan(&id, &uID, &wID, &rating, &reviewTimeStr, &responseTimeMs)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review history: %w", err)
		}

		// Parse datetime string
		reviewTime, err := r.parseDateTime(reviewTimeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse review_time: %w", err)
		}

		history := learning.NewReviewHistory(userID, wordID, learning.Rating(rating), time.Duration(responseTimeMs)*time.Millisecond)
		history.SetID(id)
		history.SetReviewTime(reviewTime)

		historyList = append(historyList, history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return historyList, nil
}

// GetUserStats retrieves learning statistics for a user
func (r *learningRepository) GetUserStats(ctx context.Context, userID user.ID) (*learning.UserStats, error) {
	stats := &learning.UserStats{}

	// Total words in vocabulary
	var totalVocabularyWords int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM words
	`).Scan(&totalVocabularyWords)
	if err != nil {
		return nil, fmt.Errorf("failed to get total vocabulary words: %w", err)
	}

	// Words that have progress records (have been studied)
	var studiedWords int
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_progress WHERE user_id = ?
	`, int64(userID)).Scan(&studiedWords)
	if err != nil {
		return nil, fmt.Errorf("failed to get studied words: %w", err)
	}

	// Calculate new words (vocabulary words minus studied words)
	newWords := totalVocabularyWords - studiedWords

	// Set totals
	stats.TotalWords = totalVocabularyWords
	stats.NewWords = newWords

	// Words by state (only for words that have been studied)
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_progress WHERE user_id = ? AND state IN ('learning', 'relearning')
	`, int64(userID)).Scan(&stats.LearningWords)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning words: %w", err)
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_progress WHERE user_id = ? AND state = 'review'
	`, int64(userID)).Scan(&stats.ReviewWords)
	if err != nil {
		return nil, fmt.Errorf("failed to get review words: %w", err)
	}

	// Due words - only count words that are actually due according to FSRS schedule
	var dueProgressWords int
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_progress WHERE user_id = ? AND due_date <= CURRENT_TIMESTAMP
	`, int64(userID)).Scan(&dueProgressWords)
	if err != nil {
		return nil, fmt.Errorf("failed to get due progress words: %w", err)
	}

	// Only count actually due words, don't artificially inflate with new words
	stats.DueWords = dueProgressWords

	// Average difficulty (only for words that have been studied)
	if studiedWords > 0 {
		err = r.db.QueryRowContext(ctx, `
			SELECT COALESCE(AVG(difficulty), 0) FROM user_progress WHERE user_id = ?
		`, int64(userID)).Scan(&stats.AvgDifficulty)
		if err != nil {
			return nil, fmt.Errorf("failed to get average difficulty: %w", err)
		}
	} else {
		stats.AvgDifficulty = 0.0
	}

	// Total reviews
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM review_history WHERE user_id = ?
	`, int64(userID)).Scan(&stats.TotalReviews)
	if err != nil {
		return nil, fmt.Errorf("failed to get total reviews: %w", err)
	}

	// Correct reviews (rating >= Good)
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM review_history WHERE user_id = ? AND rating >= 3
	`, int64(userID)).Scan(&stats.CorrectReviews)
	if err != nil {
		return nil, fmt.Errorf("failed to get correct reviews: %w", err)
	}

	return stats, nil
}

// GetUsersWithProgress retrieves all users who have learning progress
func (r *learningRepository) GetUsersWithProgress(ctx context.Context) ([]user.ID, error) {
	query := `
		SELECT DISTINCT user_id 
		FROM user_progress 
		ORDER BY user_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users with progress: %w", err)
	}
	defer rows.Close()

	var userIDs []user.ID
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, user.ID(userID))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return userIDs, nil
}

// Helper method to set FSRS card data from database values
func (r *learningRepository) setFSRSCardFromDB(card *learning.FSRSCard, stability, difficulty float64,
	lastReview, dueDate time.Time, reviewCount, lapses int, state string) {
	card.SetStability(stability)
	card.SetDifficulty(difficulty)
	card.SetLastReview(lastReview)
	card.SetDueDate(dueDate)
	card.SetReviewCount(reviewCount)
	card.SetLapses(lapses)
	card.SetState(learning.State(state))
}

// Helper method to parse datetime strings
func (r *learningRepository) parseDateTime(str sql.NullString) (time.Time, error) {
	if !str.Valid {
		return time.Time{}, nil
	}

	// Try different SQLite datetime formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02T15:04:05.000",
		"2006-01-02 15:04:05+00:00",     // SQLite with timezone
		"2006-01-02 15:04:05.000+00:00", // SQLite with milliseconds and timezone
		"2006-01-02 15:04:05-07:00",     // SQLite with different timezone
		"2006-01-02 15:04:05.000-07:00", // SQLite with milliseconds and different timezone
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str.String); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", str.String)
}

// SaveProgressAndHistory saves both progress and review history in a single transaction
func (r *learningRepository) SaveProgressAndHistory(ctx context.Context, progress *learning.UserProgress, history *learning.ReviewHistory) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Save or update progress
	fsrsCard := progress.FSRSCard()
	if progress.ID() == 0 {
		query := `
			INSERT INTO user_progress 
			(user_id, word_id, stability, difficulty, last_review, due_date, review_count, lapses, state, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		result, err := tx.ExecContext(ctx, query,
			int64(progress.UserID()), int64(progress.WordID()),
			fsrsCard.Stability(), fsrsCard.Difficulty(),
			fsrsCard.LastReview(), fsrsCard.DueDate(),
			fsrsCard.ReviewCount(), fsrsCard.Lapses(), string(fsrsCard.State()),
			progress.CreatedAt(), progress.UpdatedAt())

		if err != nil {
			return fmt.Errorf("failed to save progress: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get progress ID: %w", err)
		}
		progress.SetID(learning.ID(id))
	} else {
		query := `
			UPDATE user_progress 
			SET stability = ?, difficulty = ?, last_review = ?, due_date = ?, 
				review_count = ?, lapses = ?, state = ?, updated_at = ?
			WHERE id = ?
		`
		_, err = tx.ExecContext(ctx, query,
			fsrsCard.Stability(), fsrsCard.Difficulty(),
			fsrsCard.LastReview(), fsrsCard.DueDate(),
			fsrsCard.ReviewCount(), fsrsCard.Lapses(), string(fsrsCard.State()),
			progress.UpdatedAt(), int64(progress.ID()))

		if err != nil {
			return fmt.Errorf("failed to update progress: %w", err)
		}
	}

	// Save review history
	query := `
		INSERT INTO review_history (user_id, word_id, rating, review_time, response_time_ms)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := tx.ExecContext(ctx, query,
		int64(history.UserID()), int64(history.WordID()),
		int(history.Rating()), history.ReviewTime(), history.ResponseTimeMs())

	if err != nil {
		return fmt.Errorf("failed to save review history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get review history ID: %w", err)
	}
	history.SetID(learning.ID(id))

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
