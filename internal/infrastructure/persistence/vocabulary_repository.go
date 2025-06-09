package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"dutch-learning-bot/internal/domain/vocabulary"
)

type vocabularyRepository struct {
	db *sql.DB
}

// NewVocabularyRepository creates a new vocabulary repository
func NewVocabularyRepository(db *sql.DB) vocabulary.Repository {
	return &vocabularyRepository{db: db}
}

// Save persists a word to storage
func (r *vocabularyRepository) Save(ctx context.Context, word *vocabulary.Word) error {
	query := `
		INSERT OR IGNORE INTO words (english, dutch, category)
		VALUES (?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query, word.English(), word.Dutch(), string(word.Category()))
	if err != nil {
		return fmt.Errorf("failed to save word: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get word ID: %w", err)
	}

	if id > 0 {
		word.SetID(vocabulary.ID(id))
	}

	return nil
}

// SaveBatch persists multiple words to storage
func (r *vocabularyRepository) SaveBatch(ctx context.Context, words []*vocabulary.Word) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO words (english, dutch, category)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, word := range words {
		_, err := stmt.ExecContext(ctx, word.English(), word.Dutch(), string(word.Category()))
		if err != nil {
			return fmt.Errorf("failed to save word %s: %w", word.English(), err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID retrieves a word by its ID
func (r *vocabularyRepository) FindByID(ctx context.Context, id vocabulary.ID) (*vocabulary.Word, error) {
	query := `
		SELECT id, english, dutch, category
		FROM words WHERE id = ?
	`

	var english, dutch, category string

	err := r.db.QueryRowContext(ctx, query, int64(id)).Scan(&id, &english, &dutch, &category)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find word by ID: %w", err)
	}

	word := vocabulary.NewWord(english, dutch, vocabulary.Category(category))
	word.SetID(id)

	return word, nil
}

// FindAll retrieves all words
func (r *vocabularyRepository) FindAll(ctx context.Context) ([]*vocabulary.Word, error) {
	query := `
		SELECT id, english, dutch, category
		FROM words
		ORDER BY category, english
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query words: %w", err)
	}
	defer rows.Close()

	var words []*vocabulary.Word

	for rows.Next() {
		var id vocabulary.ID
		var english, dutch, category string

		if err := rows.Scan(&id, &english, &dutch, &category); err != nil {
			return nil, fmt.Errorf("failed to scan word: %w", err)
		}

		word := vocabulary.NewWord(english, dutch, vocabulary.Category(category))
		word.SetID(id)
		words = append(words, word)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return words, nil
}

// FindByCategory retrieves words by category
func (r *vocabularyRepository) FindByCategory(ctx context.Context, category vocabulary.Category) ([]*vocabulary.Word, error) {
	query := `
		SELECT id, english, dutch, category
		FROM words WHERE category = ?
		ORDER BY english
	`

	rows, err := r.db.QueryContext(ctx, query, string(category))
	if err != nil {
		return nil, fmt.Errorf("failed to query words by category: %w", err)
	}
	defer rows.Close()

	var words []*vocabulary.Word

	for rows.Next() {
		var id vocabulary.ID
		var english, dutch, cat string

		if err := rows.Scan(&id, &english, &dutch, &cat); err != nil {
			return nil, fmt.Errorf("failed to scan word: %w", err)
		}

		word := vocabulary.NewWord(english, dutch, vocabulary.Category(cat))
		word.SetID(id)
		words = append(words, word)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return words, nil
}

// Exists checks if a word already exists
func (r *vocabularyRepository) Exists(ctx context.Context, english, dutch string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM words
		WHERE english = ? AND dutch = ?
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, english, dutch).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check word existence: %w", err)
	}

	return count > 0, nil
}
