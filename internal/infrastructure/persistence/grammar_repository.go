package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"dutch-learning-bot/internal/domain/grammar"
)

type grammarRepository struct {
	db *sql.DB
}

// NewGrammarRepository creates a new grammar repository
func NewGrammarRepository(db *sql.DB) grammar.Repository {
	return &grammarRepository{db: db}
}

// SaveBatch saves multiple grammar tips
func (r *grammarRepository) SaveBatch(ctx context.Context, tips []*grammar.GrammarTip) error {
	for _, tip := range tips {
		query := `
			INSERT INTO grammar_tips (title, explanation, dutch_example, english_example, category, applicable_categories, word_patterns, specific_words, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		// Convert slices to JSON strings
		applicableCategoriesJSON, _ := json.Marshal(tip.ApplicableCategories())
		wordPatternsJSON, _ := json.Marshal(tip.WordPatterns())
		specificWordsJSON, _ := json.Marshal(tip.SpecificWords())

		result, err := r.db.ExecContext(ctx, query,
			tip.Title(), tip.Explanation(), tip.DutchExample(), tip.EnglishExample(),
			string(tip.Category()),
			string(applicableCategoriesJSON), string(wordPatternsJSON), string(specificWordsJSON),
			tip.CreatedAt())
		if err != nil {
			return fmt.Errorf("failed to save grammar tip %s: %w", tip.Title(), err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get grammar tip ID: %w", err)
		}

		tip.SetID(grammar.ID(id))
	}
	return nil
}

// FindApplicableToWord finds grammar tips that apply to a specific word
func (r *grammarRepository) FindApplicableToWord(ctx context.Context, dutchWord, englishWord, category string) ([]*grammar.GrammarTip, error) {
	query := `
		SELECT id, title, explanation, dutch_example, english_example, category, applicable_categories, word_patterns, specific_words, created_at
		FROM grammar_tips
		WHERE 
			JSON_EXTRACT(applicable_categories, '$') LIKE '%"' || ? || '"%' OR
			JSON_EXTRACT(specific_words, '$') LIKE '%"' || ? || '"%' OR
			JSON_EXTRACT(specific_words, '$') LIKE '%"' || ? || '"%'
		ORDER BY RANDOM()
	`

	rows, err := r.db.QueryContext(ctx, query, category, dutchWord, englishWord)
	if err != nil {
		return nil, fmt.Errorf("failed to query applicable grammar tips: %w", err)
	}
	defer rows.Close()

	var tips []*grammar.GrammarTip
	for rows.Next() {
		var id grammar.ID
		var title, explanation, dutchExample, englishExample, cat string
		var applicableCategoriesJSON, wordPatternsJSON, specificWordsJSON string
		var createdAt time.Time

		err := rows.Scan(&id, &title, &explanation, &dutchExample, &englishExample, &cat,
			&applicableCategoriesJSON, &wordPatternsJSON, &specificWordsJSON, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grammar tip: %w", err)
		}

		// Parse JSON strings back to slices
		var applicableCategories, wordPatterns, specificWords []string
		json.Unmarshal([]byte(applicableCategoriesJSON), &applicableCategories)
		json.Unmarshal([]byte(wordPatternsJSON), &wordPatterns)
		json.Unmarshal([]byte(specificWordsJSON), &specificWords)

		// Create tip and check if it actually applies (double-check with domain logic)
		tip := grammar.NewGrammarTip(
			title, explanation, dutchExample, englishExample,
			grammar.Category(cat),
			applicableCategories, wordPatterns, specificWords)
		tip.SetID(id)

		if tip.IsApplicableToWord(dutchWord, englishWord, category) {
			tips = append(tips, tip)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating grammar tips: %w", err)
	}

	return tips, nil
}
